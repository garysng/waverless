package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"waverless/internal/jobs"
	"waverless/internal/service"
	"waverless/pkg/autoscaler"
	"waverless/pkg/deploy/k8s"
	"waverless/pkg/logger"
	"waverless/pkg/monitoring"
	mysqlstore "waverless/pkg/store/mysql"
)

func (app *Application) initJobs() error {
	if app.workerService == nil || app.taskService == nil {
		logger.WarnCtx(app.ctx, "Service layer not fully initialized yet, skipping background task registration")
		return nil
	}

	manager := jobs.NewManager(app.ctx)

	workerInterval := time.Duration(app.config.Worker.HeartbeatTimeout) * time.Second
	if workerInterval <= 0 {
		workerInterval = time.Minute
	}

	// Create distributed locks to prevent multiple replicas from executing background cleanup tasks simultaneously
	// If Redis is unavailable, locks will automatically downgrade to single-instance mode
	var redisClient *redis.Client
	if app.redisClient != nil {
		redisClient = app.redisClient.GetClient()
	}

	workerCleanupLock := autoscaler.NewRedisDistributedLock(redisClient, "cleanup:worker-lock")
	taskTimeoutLock := autoscaler.NewRedisDistributedLock(redisClient, "cleanup:task-timeout-lock")
	orphanedTaskLock := autoscaler.NewRedisDistributedLock(redisClient, "cleanup:orphaned-task-lock")

	// Register background tasks with locks
	manager.Register(newWorkerCleanupJob(workerInterval, app.workerService, workerCleanupLock))
	manager.Register(newTaskTimeoutCleanupJob(5*time.Minute, app.taskService, taskTimeoutLock))
	manager.Register(newOrphanedTaskCleanupJob(15*time.Second, app.taskService, orphanedTaskLock))

	// Register task statistics refresh task
	if app.statisticsService != nil {
		statsRefreshLock := autoscaler.NewRedisDistributedLock(redisClient, "stats:refresh-lock")
		manager.Register(newStatisticsRefreshJob(10*time.Minute, app.statisticsService, statsRefreshLock))
	}

	// Register monitoring tasks
	if app.monitoringService != nil {
		minuteAggLock := autoscaler.NewRedisDistributedLock(redisClient, "monitoring:minute-agg-lock")
		hourlyAggLock := autoscaler.NewRedisDistributedLock(redisClient, "monitoring:hourly-agg-lock")
		dailyAggLock := autoscaler.NewRedisDistributedLock(redisClient, "monitoring:daily-agg-lock")
		snapshotLock := autoscaler.NewRedisDistributedLock(redisClient, "monitoring:snapshot-lock")
		dataCleanupLock := autoscaler.NewRedisDistributedLock(redisClient, "cleanup:data-retention-lock")

		manager.Register(newMinuteAggregationJob(time.Minute, app.monitoringService, minuteAggLock))
		manager.Register(newHourlyAggregationJob(time.Hour, app.monitoringService, hourlyAggLock))
		manager.Register(newDailyAggregationJob(24*time.Hour, app.monitoringService, dailyAggLock))
		manager.Register(newSnapshotCollectionJob(time.Minute, app.monitoringCollector, snapshotLock))
		manager.Register(newDataRetentionCleanupJob(24*time.Hour, app.mysqlRepo, dataCleanupLock))
	}

	app.jobsManager = manager
	return nil
}

// workerCleanupJob periodically removes offline workers.
type workerCleanupJob struct {
	interval        time.Duration
	workerService   *service.WorkerService
	distributedLock autoscaler.DistributedLock
}

func newWorkerCleanupJob(interval time.Duration, svc *service.WorkerService, lock autoscaler.DistributedLock) jobs.Job {
	return &workerCleanupJob{
		interval:        interval,
		workerService:   svc,
		distributedLock: lock,
	}
}

func (j *workerCleanupJob) Name() string {
	return "worker-cleanup"
}

func (j *workerCleanupJob) Interval() time.Duration {
	return j.interval
}

func (j *workerCleanupJob) Run(ctx context.Context) error {
	if j.workerService == nil {
		return fmt.Errorf("worker service not configured")
	}

	// Try to acquire distributed lock
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			logger.DebugCtx(ctx, "another instance is running worker cleanup, skipping this cycle")
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}

	logger.DebugCtx(ctx, "running worker cleanup job")
	return j.workerService.CleanupOfflineWorkers(ctx)
}

// taskTimeoutCleanupJob reconciles timed-out tasks as a fallback mechanism.
type taskTimeoutCleanupJob struct {
	interval        time.Duration
	taskService     *service.TaskService
	distributedLock autoscaler.DistributedLock
}

func newTaskTimeoutCleanupJob(interval time.Duration, svc *service.TaskService, lock autoscaler.DistributedLock) jobs.Job {
	return &taskTimeoutCleanupJob{
		interval:        interval,
		taskService:     svc,
		distributedLock: lock,
	}
}

func (j *taskTimeoutCleanupJob) Name() string {
	return "task-timeout-cleanup"
}

func (j *taskTimeoutCleanupJob) Interval() time.Duration {
	return j.interval
}

func (j *taskTimeoutCleanupJob) Run(ctx context.Context) error {
	if j.taskService == nil {
		return fmt.Errorf("task service not configured")
	}

	// Try to acquire distributed lock
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			logger.DebugCtx(ctx, "another instance is running task timeout cleanup, skipping this cycle")
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}

	logger.DebugCtx(ctx, "running task timeout cleanup job")
	return j.taskService.CleanupTimedOutTasks(ctx)
}

// orphanedTaskCleanupJob reclaims tasks assigned to nonexistent workers.
type orphanedTaskCleanupJob struct {
	interval        time.Duration
	taskService     *service.TaskService
	distributedLock autoscaler.DistributedLock
}

func newOrphanedTaskCleanupJob(interval time.Duration, svc *service.TaskService, lock autoscaler.DistributedLock) jobs.Job {
	return &orphanedTaskCleanupJob{
		interval:        interval,
		taskService:     svc,
		distributedLock: lock,
	}
}

func (j *orphanedTaskCleanupJob) Name() string {
	return "orphaned-task-cleanup"
}

func (j *orphanedTaskCleanupJob) Interval() time.Duration {
	return j.interval
}

func (j *orphanedTaskCleanupJob) Run(ctx context.Context) error {
	if j.taskService == nil {
		return fmt.Errorf("task service not configured")
	}

	// Try to acquire distributed lock
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			logger.DebugCtx(ctx, "another instance is running orphaned task cleanup, skipping this cycle")
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}

	logger.DebugCtx(ctx, "running orphaned task cleanup job")
	return j.taskService.CleanupOrphanedTasks(ctx)
}

// startPodCleanupJob starts Pod cleanup task (handles stuck Terminating Pods)
func (app *Application) startPodCleanupJob(k8sProvider *k8s.K8sDeploymentProvider) error {
	if k8sProvider == nil {
		logger.InfoCtx(app.ctx, "K8s provider not available, skipping pod cleanup job")
		return nil
	}

	logger.InfoCtx(app.ctx, "Starting pod cleanup job for stuck terminating pods...")

	// Start background task, check every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-app.ctx.Done():
				logger.InfoCtx(app.ctx, "Pod cleanup job stopped")
				return
			case <-ticker.C:
				app.cleanupStuckTerminatingPods(k8sProvider)
			}
		}
	}()

	logger.InfoCtx(app.ctx, "✅ Pod cleanup job started")
	return nil
}

// cleanupStuckTerminatingPods cleans up stuck Terminating Pods
func (app *Application) cleanupStuckTerminatingPods(k8sProvider *k8s.K8sDeploymentProvider) {
	ctx := app.ctx

	// Get all Pods (including Terminating)
	pods, err := k8sProvider.GetPods(ctx, "")
	if err != nil {
		logger.WarnCtx(ctx, "Failed to get pods for cleanup check: %v", err)
		return
	}

	for _, pod := range pods {
		// Only process Terminating Pods
		if pod.DeletionTimestamp == "" {
			continue
		}

		// Check if Worker has running tasks (podName = workerID)
		workerID := pod.Name
		hasRunningTasks, err := app.hasRunningTasks(ctx, workerID)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to check running tasks for worker %s: %v, skipping force delete", workerID, err)
			continue
		}

		if hasRunningTasks {
			logger.InfoCtx(ctx, "Worker %s has running tasks, NOT force deleting pod (waiting for tasks to complete)", workerID)
			continue
		}

		// ✅ Worker has no running tasks and pod is Terminating - kill main process
		logger.WarnCtx(ctx, "🔥 Pod %s is stuck in Terminating with NO running tasks, killing main process (PID 1)...",
			pod.Name)

		// Get endpoint from pod labels (pod.Labels["app"] = endpoint)
		endpoint := ""
		if pod.Labels != nil {
			endpoint = pod.Labels["app"]
		}
		if endpoint == "" {
			logger.ErrorCtx(ctx, "Failed to get endpoint from pod %s labels, skipping", pod.Name)
			continue
		}

		// Execute pkill -9 1 to kill the main process (PID 1) in container
		// This will cause container to exit and pod to terminate naturally
		stdout, stderr, err := k8sProvider.ExecPodCommand(ctx, pod.Name, endpoint, []string{"sh", "-c", "pkill -9 1"})
		if err != nil {
			logger.ErrorCtx(ctx, "Failed to kill main process in pod %s: %v, stdout: %s, stderr: %s",
				pod.Name, err, stdout, stderr)
		} else {
			logger.InfoCtx(ctx, "✅ Successfully killed main process in pod %s (stdout: %s, stderr: %s)",
				pod.Name, stdout, stderr)
		}
	}
}

// hasRunningTasks checks if worker has running tasks (via database)
func (app *Application) hasRunningTasks(ctx context.Context, workerID string) (bool, error) {
	// Query database directly via task repository
	tasks, err := app.mysqlRepo.Task.GetTasksByWorker(ctx, workerID)
	if err != nil {
		return false, fmt.Errorf("failed to get tasks for worker %s: %w", workerID, err)
	}

	// GetTasksByWorker 只返回 IN_PROGRESS 状态的任务
	return len(tasks) > 0, nil
}

// statisticsRefreshJob periodically refreshes task statistics from the tasks table
type statisticsRefreshJob struct {
	interval          time.Duration
	statisticsService *service.StatisticsService
	distributedLock   autoscaler.DistributedLock
}

func newStatisticsRefreshJob(interval time.Duration, svc *service.StatisticsService, lock autoscaler.DistributedLock) jobs.Job {
	return &statisticsRefreshJob{
		interval:          interval,
		statisticsService: svc,
		distributedLock:   lock,
	}
}

func (j *statisticsRefreshJob) Name() string {
	return "statistics-refresh"
}

func (j *statisticsRefreshJob) Interval() time.Duration {
	return j.interval
}

func (j *statisticsRefreshJob) Run(ctx context.Context) error {
	if j.statisticsService == nil {
		return fmt.Errorf("statistics service not configured")
	}

	// Try to acquire distributed lock
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			logger.DebugCtx(ctx, "another instance is running statistics refresh, skipping this cycle")
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}

	logger.InfoCtx(ctx, "running statistics refresh job")
	return j.statisticsService.RefreshAllStatistics(ctx)
}


// minuteAggregationJob aggregates monitoring data every minute
type minuteAggregationJob struct {
	interval          time.Duration
	monitoringService *service.MonitoringService
	distributedLock   autoscaler.DistributedLock
}

func newMinuteAggregationJob(interval time.Duration, svc *service.MonitoringService, lock autoscaler.DistributedLock) jobs.Job {
	return &minuteAggregationJob{
		interval:          interval,
		monitoringService: svc,
		distributedLock:   lock,
	}
}

func (j *minuteAggregationJob) Name() string { return "monitoring-minute-aggregation" }

func (j *minuteAggregationJob) Interval() time.Duration { return j.interval }

func (j *minuteAggregationJob) Run(ctx context.Context) error {
	if j.monitoringService == nil {
		return fmt.Errorf("monitoring service not configured")
	}
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}
	return j.monitoringService.AggregateMinuteStats(ctx)
}

// hourlyAggregationJob aggregates monitoring data every hour
type hourlyAggregationJob struct {
	interval          time.Duration
	monitoringService *service.MonitoringService
	distributedLock   autoscaler.DistributedLock
}

func newHourlyAggregationJob(interval time.Duration, svc *service.MonitoringService, lock autoscaler.DistributedLock) jobs.Job {
	return &hourlyAggregationJob{
		interval:          interval,
		monitoringService: svc,
		distributedLock:   lock,
	}
}

func (j *hourlyAggregationJob) Name() string { return "monitoring-hourly-aggregation" }

func (j *hourlyAggregationJob) Interval() time.Duration { return j.interval }

func (j *hourlyAggregationJob) AlignToInterval() bool { return true }

func (j *hourlyAggregationJob) Run(ctx context.Context) error {
	if j.monitoringService == nil {
		return fmt.Errorf("monitoring service not configured")
	}
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}
	return j.monitoringService.AggregateHourlyStats(ctx)
}

// dailyAggregationJob aggregates monitoring data every day
type dailyAggregationJob struct {
	interval          time.Duration
	monitoringService *service.MonitoringService
	distributedLock   autoscaler.DistributedLock
}

func newDailyAggregationJob(interval time.Duration, svc *service.MonitoringService, lock autoscaler.DistributedLock) jobs.Job {
	return &dailyAggregationJob{
		interval:          interval,
		monitoringService: svc,
		distributedLock:   lock,
	}
}

func (j *dailyAggregationJob) Name() string { return "monitoring-daily-aggregation" }

func (j *dailyAggregationJob) Interval() time.Duration { return j.interval }

func (j *dailyAggregationJob) AlignToInterval() bool { return true }

func (j *dailyAggregationJob) Run(ctx context.Context) error {
	if j.monitoringService == nil {
		return fmt.Errorf("monitoring service not configured")
	}
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}
	return j.monitoringService.AggregateDailyStats(ctx)
}

// snapshotCollectionJob collects worker resource snapshots
type snapshotCollectionJob struct {
	interval        time.Duration
	collector       *monitoring.Collector
	distributedLock autoscaler.DistributedLock
}

func newSnapshotCollectionJob(interval time.Duration, collector *monitoring.Collector, lock autoscaler.DistributedLock) jobs.Job {
	return &snapshotCollectionJob{interval: interval, collector: collector, distributedLock: lock}
}

func (j *snapshotCollectionJob) Name() string { return "monitoring-snapshot-collection" }

func (j *snapshotCollectionJob) Interval() time.Duration { return j.interval }

func (j *snapshotCollectionJob) Run(ctx context.Context) error {
	if j.collector == nil {
		return nil
	}
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}
	return j.collector.CollectSnapshots(ctx)
}


// dataRetentionCleanupJob cleans up old data (tasks, task_events, worker_events) daily
type dataRetentionCleanupJob struct {
	interval        time.Duration
	repo            *mysqlstore.Repository
	distributedLock autoscaler.DistributedLock
}

func newDataRetentionCleanupJob(interval time.Duration, repo *mysqlstore.Repository, lock autoscaler.DistributedLock) jobs.Job {
	return &dataRetentionCleanupJob{interval: interval, repo: repo, distributedLock: lock}
}

func (j *dataRetentionCleanupJob) Name() string { return "data-retention-cleanup" }

func (j *dataRetentionCleanupJob) Interval() time.Duration { return j.interval }

func (j *dataRetentionCleanupJob) Run(ctx context.Context) error {
	if j.repo == nil {
		return nil
	}
	if j.distributedLock != nil {
		acquired, err := j.distributedLock.TryLock(ctx)
		if err != nil || !acquired {
			return nil
		}
		defer j.distributedLock.Unlock(ctx)
	}

	retentionDays := 10
	before := time.Now().AddDate(0, 0, -retentionDays)
	
	// Clean old completed/failed tasks
	taskRows, _ := j.repo.Task.CleanupOldTasks(ctx, before)
	if taskRows > 0 {
		logger.InfoCtx(ctx, "cleaned up %d old tasks (older than %d days)", taskRows, retentionDays)
	}

	// Clean old task events
	eventRows, _ := j.repo.TaskEvent.CleanupOldEvents(ctx, before)
	if eventRows > 0 {
		logger.InfoCtx(ctx, "cleaned up %d old task events (older than %d days)", eventRows, retentionDays)
	}

	// Clean old worker events
	workerEventRows, _ := j.repo.Monitoring.CleanupOldWorkerEvents(ctx, before)
	if workerEventRows > 0 {
		logger.InfoCtx(ctx, "cleaned up %d old worker events (older than %d days)", workerEventRows, retentionDays)
	}

	return nil
}
