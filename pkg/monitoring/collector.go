package monitoring

import (
	"context"
	"time"

	"waverless/pkg/logger"
	"waverless/pkg/store/mysql"
	"waverless/pkg/store/mysql/model"
)

// Collector collects worker resource snapshots
type Collector struct {
	repo       *mysql.MonitoringRepository
	workerRepo *mysql.WorkerRepository
	taskRepo   *mysql.TaskRepository
}

// NewCollector creates a new resource collector
func NewCollector(repo *mysql.MonitoringRepository, workerRepo *mysql.WorkerRepository, taskRepo *mysql.TaskRepository) *Collector {
	return &Collector{repo: repo, workerRepo: workerRepo, taskRepo: taskRepo}
}

// CollectSnapshots collects resource snapshots for all active workers
func (c *Collector) CollectSnapshots(ctx context.Context) error {
	workers, err := c.workerRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, w := range workers {
		// Get current task for this worker
		tasks, _ := c.taskRepo.GetTasksByWorker(ctx, w.WorkerID)
		var currentTaskID string
		if len(tasks) > 0 {
			currentTaskID = tasks[0].TaskID
		}

		snapshot := &model.WorkerResourceSnapshot{
			WorkerID:      w.WorkerID,
			Endpoint:      w.Endpoint, // 填充 endpoint 字段
			SnapshotAt:    now,
			IsIdle:        currentTaskID == "",
			CurrentTaskID: currentTaskID,
			// GPU/CPU metrics would come from K8s metrics API in production
			// For now, we just track idle state
		}

		if err := c.repo.SaveResourceSnapshot(ctx, snapshot); err != nil {
			logger.ErrorCtx(ctx, "failed to save snapshot for worker %s: %v", w.WorkerID, err)
		}
	}

	// Cleanup old snapshots (keep 24 hours)
	c.repo.CleanupOldSnapshots(ctx, now.Add(-24*time.Hour))
	return nil
}
