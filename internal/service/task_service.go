package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"waverless/internal/model"
	endpointsvc "waverless/internal/service/endpoint"
	"waverless/pkg/config"
	"waverless/pkg/interfaces"
	"waverless/pkg/logger"
	"waverless/pkg/store/mysql"
	mysqlModel "waverless/pkg/store/mysql/model"

	"github.com/google/uuid"
)

// TaskService Task service
type TaskService struct {
	taskRepo           *mysql.TaskRepository
	taskEventRepo      *mysql.TaskEventRepository
	endpointService    *endpointsvc.Service
	deploymentProvider interfaces.DeploymentProvider
	statisticsService  *StatisticsService
	workerService      *WorkerService
}

// NewTaskService creates a new Task service
func NewTaskService(taskRepo *mysql.TaskRepository, taskEventRepo *mysql.TaskEventRepository, endpointService *endpointsvc.Service, deploymentProvider interfaces.DeploymentProvider) *TaskService {
	return &TaskService{
		taskRepo:           taskRepo,
		taskEventRepo:      taskEventRepo,
		endpointService:    endpointService,
		deploymentProvider: deploymentProvider,
	}
}

// SetStatisticsService sets the statistics service (for dependency injection)
func (s *TaskService) SetStatisticsService(statsService *StatisticsService) {
	s.statisticsService = statsService
}

// SetWorkerService sets the worker service (for dependency injection)
func (s *TaskService) SetWorkerService(workerService *WorkerService) {
	s.workerService = workerService
}

// SubmitTask submits a task
func (s *TaskService) SubmitTask(ctx context.Context, req *model.SubmitRequest) (*model.SubmitResponse, error) {
	taskID := uuid.New().String()

	// Process endpoint
	endpoint := req.Endpoint
	if endpoint == "" {
		endpoint = "default"
	}

	task := &model.Task{
		ID:         taskID,
		Endpoint:   endpoint,
		Input:      req.Input,
		Status:     model.TaskStatusPending,
		WebhookURL: req.WebhookURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save task to MySQL
	mysqlTask := mysql.FromTaskDomain(task)
	if err := s.taskRepo.Create(ctx, mysqlTask); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// Record TASK_CREATED event (task is in MySQL with PENDING status, worker will pull directly from database)
	s.recordTaskCreated(ctx, mysqlTask)

	// Record TASK_QUEUED event
	s.recordTaskQueued(ctx, mysqlTask)

	// Save updated extend field
	if err := s.taskRepo.UpdateFields(ctx, mysqlTask.TaskID, map[string]interface{}{
		"extend": mysqlTask.Extend,
	}); err != nil {
		logger.ErrorCtx(ctx, "failed to update task extend: %v", err)
	}

	// Asynchronously update statistics (new task: fromStatus="", toStatus=PENDING)
	if s.statisticsService != nil {
		go s.statisticsService.UpdateStatisticsOnTaskStatusChange(context.Background(), endpoint, "", "PENDING")
	}

	logger.InfoCtx(ctx, "task submitted, task_id: %s, endpoint: %s", taskID, endpoint)

	return &model.SubmitResponse{
		ID:     taskID,
		Status: model.TaskStatusPending,
	}, nil
}

// SubmitTaskSync submits a task synchronously (with timeout waiting)
func (s *TaskService) SubmitTaskSync(ctx context.Context, req *model.SubmitRequest, timeout time.Duration) (*model.TaskResponse, error) {
	resp, err := s.SubmitTask(ctx, req)
	if err != nil {
		return nil, err
	}

	// Poll and wait for task completion
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeoutTimer.C:
			return nil, fmt.Errorf("task timeout")
		case <-ticker.C:
			mysqlTask, err := s.taskRepo.Get(ctx, resp.ID)
			if err != nil {
				continue
			}

			task := mysql.ToTaskDomain(mysqlTask)
			if task.Status == model.TaskStatusCompleted || task.Status == model.TaskStatusFailed {
				return s.toTaskResponse(task), nil
			}
		}
	}
}

// GetTaskStatus gets task status
func (s *TaskService) GetTaskStatus(ctx context.Context, taskID string) (*model.TaskResponse, error) {
	mysqlTask, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task := mysql.ToTaskDomain(mysqlTask)
	return s.toTaskResponse(task), nil
}

// toTaskResponse converts Task to RunPod format response
func (s *TaskService) toTaskResponse(task *model.Task) *model.TaskResponse {
	// Calculate delay and execution time (in milliseconds)
	var delayTime int64 = 0
	var executionTime int64 = 0

	if task.StartedAt != nil {
		// Processing delay = start time - creation time
		delayTime = task.StartedAt.Sub(task.CreatedAt).Milliseconds()
	}

	if task.CompletedAt != nil && task.StartedAt != nil {
		// Execution time = completion time - start time
		executionTime = task.CompletedAt.Sub(*task.StartedAt).Milliseconds()
	}

	return &model.TaskResponse{
		ID:          task.ID,
		Status:      string(task.Status),
		Endpoint:    task.Endpoint,
		WorkerID:    task.WorkerID,
		DelayTime:   delayTime,
		ExecutionMS: executionTime,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339), // ISO 8601 format
		Input:       task.Input,
		Output:      task.Output,
		Error:       task.Error,
	}
}

// CancelTask cancels a task
func (s *TaskService) CancelTask(ctx context.Context, taskID string) error {
	mysqlTask, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if mysqlTask == nil {
		return fmt.Errorf("task not found")
	}

	if mysqlTask.Status == string(model.TaskStatusCompleted) || mysqlTask.Status == string(model.TaskStatusFailed) {
		return fmt.Errorf("task already finished")
	}

	oldStatus := mysqlTask.Status
	endpoint := mysqlTask.Endpoint
	mysqlTask.Status = string(model.TaskStatusCancelled)
	now := time.Now()
	mysqlTask.UpdatedAt = now
	mysqlTask.CompletedAt = nil
	mysqlTask.StartedAt = nil
	mysqlTask.WorkerID = ""

	updateFields := map[string]interface{}{
		"status":       mysqlTask.Status,
		"updated_at":   mysqlTask.UpdatedAt,
		"completed_at": mysqlTask.CompletedAt,
		"started_at":   mysqlTask.StartedAt,
		"worker_id":    mysqlTask.WorkerID,
	}

	if err := s.taskRepo.UpdateFields(ctx, mysqlTask.TaskID, updateFields); err != nil {
		return err
	}

	// Asynchronously update statistics (PENDING/IN_PROGRESS -> CANCELLED)
	if s.statisticsService != nil {
		go s.statisticsService.UpdateStatisticsOnTaskStatusChange(context.Background(), endpoint, oldStatus, mysqlTask.Status)
	}

	logger.InfoCtx(ctx, "task cancelled, task_id: %s", taskID)
	return nil
}

// UpdateTaskResult updates task result
func (s *TaskService) UpdateTaskResult(ctx context.Context, req *model.JobResultRequest) error {
	mysqlTask, err := s.taskRepo.Get(ctx, req.TaskID)
	if err != nil {
		return err
	}

	now := time.Now()
	oldStatus := mysqlTask.Status // Save original status for statistics
	endpoint := mysqlTask.Endpoint

	// Prepare update fields
	updates := map[string]interface{}{
		"updated_at":   now,
		"completed_at": now,
	}

	var newStatus string
	if req.Error != "" {
		newStatus = "FAILED"
		updates["status"] = newStatus
		updates["error"] = req.Error

		// Record TASK_FAILED event and update extend
		mysqlTask.Status = newStatus
		s.recordTaskFailed(ctx, mysqlTask, mysqlTask.WorkerID, req.Error)
		updates["extend"] = mysqlTask.Extend
	} else {
		newStatus = "COMPLETED"
		updates["status"] = newStatus
		updates["output"] = mysql.JSONMap(req.Output)

		// Record TASK_COMPLETED event and update extend
		mysqlTask.Status = newStatus
		mysqlTask.Output = mysql.JSONMap(req.Output)
		s.recordTaskCompleted(ctx, mysqlTask, mysqlTask.WorkerID)
		updates["extend"] = mysqlTask.Extend
	}

	// Update directly with WHERE + Updates
	err = s.taskRepo.UpdateFields(ctx, req.TaskID, updates)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Asynchronously update statistics
	if s.statisticsService != nil {
		go s.statisticsService.UpdateStatisticsOnTaskStatusChange(context.Background(), endpoint, oldStatus, newStatus)
	}

	// Record worker task completion stats
	if s.workerService != nil && mysqlTask.WorkerID != "" {
		executionMs := int64(0)
		if mysqlTask.StartedAt != nil {
			executionMs = now.Sub(*mysqlTask.StartedAt).Milliseconds()
		}
		go s.workerService.RecordTaskCompletion(context.Background(), mysqlTask.WorkerID, newStatus == "COMPLETED", executionMs)
	}

	logger.InfoCtx(ctx, "task result updated, task_id: %s, status: %s", req.TaskID, updates["status"])

	// 🔥 CRITICAL: Update mysqlTask.CompletedAt before recording GPU usage
	// The mysqlTask object was fetched from DB before updates, so CompletedAt is still nil
	// We must update it with the value from updates map for GPU usage recording to work
	mysqlTask.CompletedAt = &now

	// 🔥 CRITICAL: Update endpoint's LastTaskTime (for autoscaler idle time calculation)
	// If not updated, autoscaler will think endpoint is always idle, causing immediate scale-down after task completion
	if mysqlTask.Endpoint != "" {
		endpoint, err := s.endpointService.GetEndpoint(ctx, mysqlTask.Endpoint)
		if err != nil {
			logger.WarnCtx(ctx, "failed to get endpoint %s for updating LastTaskTime: %v", mysqlTask.Endpoint, err)
		} else {
			endpoint.LastTaskTime = now
			if updateErr := s.endpointService.UpdateEndpoint(ctx, endpoint); updateErr != nil {
				logger.WarnCtx(ctx, "failed to update endpoint %s LastTaskTime: %v", mysqlTask.Endpoint, updateErr)
			} else {
				// Use INFO level to ensure log visibility
				logger.InfoCtx(ctx, "✅ Updated endpoint %s LastTaskTime to %s (for autoscaler idle time tracking)",
					mysqlTask.Endpoint, now.Format(time.RFC3339))
			}
		}
	}

	// If webhook is configured, call asynchronously
	if mysqlTask.WebhookURL != "" {
		task := mysql.ToTaskDomain(mysqlTask)
		task.Status = model.TaskStatus(updates["status"].(string))
		task.UpdatedAt = now
		task.CompletedAt = &now
		if req.Error != "" {
			task.Error = req.Error
		} else {
			task.Output = req.Output
		}
		go s.callWebhook(context.Background(), task)
	}

	return nil
}

// callWebhook calls webhook callback
func (s *TaskService) callWebhook(ctx context.Context, task *model.Task) {
	// Build callback payload (RunPod format compatible)
	payload := s.toTaskResponse(task)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.ErrorCtx(ctx, "failed to marshal webhook payload, task_id: %s, error: %v", task.ID, err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", task.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.ErrorCtx(ctx, "failed to create webhook request, task_id: %s, error: %v", task.ID, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Waverless/1.0")

	// Send request (with timeout)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorCtx(ctx, "failed to call webhook, task_id: %s, url: %s, error: %v", task.ID, task.WebhookURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.InfoCtx(ctx, "webhook called successfully, task_id: %s, url: %s, status_code: %d", task.ID, task.WebhookURL, resp.StatusCode)
	} else {
		logger.WarnCtx(ctx, "webhook returned non-2xx status, task_id: %s, url: %s, status_code: %d", task.ID, task.WebhookURL, resp.StatusCode)
	}
}

// GetPendingTaskCount gets pending task count (by endpoint) - from MySQL statistics
func (s *TaskService) GetPendingTaskCount(ctx context.Context, endpoint string) (int64, error) {
	if endpoint == "" {
		endpoint = "default"
	}
	return s.taskRepo.CountByEndpointAndStatus(ctx, endpoint, string(model.TaskStatusPending))
}

// CheckSubmitEligibility checks if task submission is recommended based on pending queue depth
func (s *TaskService) CheckSubmitEligibility(ctx context.Context, endpoint string) (bool, int64, int, error) {
	if endpoint == "" {
		endpoint = "default"
	}

	// Get endpoint configuration
	endpointMeta, err := s.endpointService.GetEndpoint(ctx, endpoint)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to get endpoint: %w", err)
	}
	if endpointMeta == nil {
		return false, 0, 0, fmt.Errorf("endpoint not found: %s", endpoint)
	}

	// Get current pending task count
	pendingCount, err := s.GetPendingTaskCount(ctx, endpoint)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to get pending task count: %w", err)
	}

	// Get max pending tasks threshold (default to 1 if not set)
	maxPendingTasks := endpointMeta.MaxPendingTasks
	if maxPendingTasks <= 0 {
		maxPendingTasks = 1
	}

	// Check if submission is recommended
	shouldSubmit := pendingCount < int64(maxPendingTasks)

	return shouldSubmit, pendingCount, maxPendingTasks, nil
}

// ListTasks retrieves a list of tasks with optional filtering
// OPTIMIZATION: Excludes input field to avoid fetching potentially large data (e.g., base64 images)
func (s *TaskService) ListTasks(ctx context.Context, status string, endpoint string, taskID string, limit int, offset int) ([]*model.TaskResponse, int64, error) {
	// Build filters
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	if endpoint != "" {
		filters["endpoint"] = endpoint
	}

	// Get total count with same filters
	total, err := s.taskRepo.CountWithTaskID(ctx, filters, taskID)
	if err != nil {
		return nil, 0, err
	}

	// Use the optimized List method that excludes input field
	mysqlTasks, err := s.taskRepo.ListWithTaskIDExcludeInput(ctx, filters, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert to TaskResponse format (input will be nil/empty)
	responses := make([]*model.TaskResponse, 0, len(mysqlTasks))
	for _, mysqlTask := range mysqlTasks {
		task := mysql.ToTaskDomain(mysqlTask)
		responses = append(responses, s.toTaskResponse(task))
	}

	return responses, total, nil
}

// CountTasksByStatus counts tasks by status globally
func (s *TaskService) CountTasksByStatus(ctx context.Context, status string) (int64, error) {
	return s.taskRepo.CountByStatus(ctx, status)
}

// CleanupOrphanedTasks checks for tasks assigned to workers that no longer exist
// This handles cases where workers crash or are scaled down while tasks are in progress
func (s *TaskService) CleanupOrphanedTasks(ctx context.Context) error {
	logger.DebugCtx(ctx, "starting orphaned task cleanup")

	// Get all in-progress task IDs
	taskIDs, err := s.taskRepo.GetInProgressTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	if len(taskIDs) == 0 {
		logger.DebugCtx(ctx, "no in-progress tasks to check for orphans")
		return nil
	}

	logger.DebugCtx(ctx, "checking %d in-progress tasks for orphaned workers", len(taskIDs))

	orphanedCount := 0

	// Check each in-progress task
	for _, taskID := range taskIDs {
		mysqlTask, err := s.taskRepo.Get(ctx, taskID)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to get task during orphan check, task_id: %s, error: %v", taskID, err)
			continue
		}
		if mysqlTask == nil {
			continue
		}

		// Skip if task is no longer in progress (race condition)
		if mysqlTask.Status != string(model.TaskStatusInProgress) {
			continue
		}

		// Check if task has a worker assigned
		if mysqlTask.WorkerID == "" {
			logger.WarnCtx(ctx, "orphaned task detected (no worker assigned), task_id: %s, endpoint: %s",
				taskID, mysqlTask.Endpoint)
			s.requeueOrphanedTask(ctx, mysqlTask, "No worker assigned")
			orphanedCount++
			continue
		}

		// Check if the assigned worker exists in database (any status except OFFLINE)
		if s.workerService == nil {
			logger.ErrorCtx(ctx, "orphaned workerService is nil, cannot check orphaned tasks")
			continue
		}
		worker, err := s.workerService.GetWorker(ctx, mysqlTask.WorkerID)
		if err != nil {
			logger.WarnCtx(ctx, "orphaned task: failed to get worker %s: %v", mysqlTask.WorkerID, err)
		}
		if err != nil || worker == nil || worker.Status == model.WorkerStatusOffline {
			logger.WarnCtx(ctx, "orphaned task detected (worker offline/not found), task_id: %s, worker_id: %s, endpoint: %s, worker_status: %v",
				taskID, mysqlTask.WorkerID, mysqlTask.Endpoint, worker)
			s.recordTaskOrphaned(ctx, mysqlTask)
			s.requeueOrphanedTask(ctx, mysqlTask, fmt.Sprintf("Worker %s no longer exists", mysqlTask.WorkerID))
			orphanedCount++
		}
	}

	if orphanedCount > 0 {
		logger.InfoCtx(ctx, "orphaned task cleanup completed, orphaned: %d, total_checked: %d", orphanedCount, len(taskIDs))
	}

	return nil
}

// requeueOrphanedTask re-queues an orphaned task for retry
// Unlike timeout/failure, orphaned tasks are healthy but lost their worker (crash/scale-down)
// They should be given another chance to execute
//
// OPTIMIZATION: Uses transaction to ensure atomicity and reduce database operations from 3 to 2
func (s *TaskService) requeueOrphanedTask(ctx context.Context, task *mysql.Task, reason string) {
	logger.InfoCtx(ctx, "re-queuing orphaned task, task_id: %s, endpoint: %s, reason: %s",
		task.TaskID, task.Endpoint, reason)

	now := time.Now()
	oldStatus := task.Status

	// Prepare status updates
	task.Status = string(model.TaskStatusPending)
	task.WorkerID = ""
	task.StartedAt = nil
	task.CompletedAt = nil
	task.UpdatedAt = now

	// Update extend field for TASK_REQUEUED event (in-memory)
	// Note: We use updateTaskExtend instead of recordTaskRequeued to avoid async goroutine
	// which cannot be rolled back in transaction
	s.updateTaskExtend(task, mysqlModel.EventTaskRequeued, "")

	// 🔒 Use transaction to execute all database operations atomically
	err := s.taskRepo.ExecTx(ctx, func(txCtx context.Context) error {
		// 1️⃣ Update all task fields in a single UPDATE (including extend)
		// ⚠️ CRITICAL: Use CAS to prevent race condition with task completion
		// Only update if task is still IN_PROGRESS (may have been completed by worker in the meantime)
		fieldUpdates := map[string]interface{}{
			"status":       task.Status,
			"worker_id":    task.WorkerID,
			"started_at":   task.StartedAt,
			"completed_at": task.CompletedAt,
			"updated_at":   task.UpdatedAt,
			"extend":       task.Extend, // ✅ Merged: avoid second UPDATE
		}

		if err := s.taskRepo.UpdateFieldsWithStatus(txCtx, task.TaskID, "IN_PROGRESS", fieldUpdates); err != nil {
			// Check if error is due to status change (task completed/cancelled by worker)
			if strings.Contains(err.Error(), "status changed") {
				logger.InfoCtx(txCtx, "⏭️  Task %s status changed during requeue (likely completed by worker), skipping requeue",
					task.TaskID)
				return nil // Not an error, just skip this task
			}
			return fmt.Errorf("failed to update orphaned task: %w", err)
		}

		// 2️⃣ Record TASK_REQUEUED event (synchronous, within transaction)
		event := &mysqlModel.TaskEvent{
			TaskID:       task.TaskID,
			Endpoint:     task.Endpoint,
			EventType:    string(mysqlModel.EventTaskRequeued),
			EventTime:    now,
			WorkerID:     "",
			FromStatus:   oldStatus,
			ErrorMessage: reason,
		}

		if err := s.taskEventRepo.RecordEvent(txCtx, event); err != nil {
			return fmt.Errorf("failed to record task event: %w", err)
		}

		return nil // ✅ Commit transaction
	})

	if err != nil {
		logger.ErrorCtx(ctx, "failed to requeue orphaned task (transaction rolled back), task_id: %s, error: %v",
			task.TaskID, err)
		return
	}

	// Asynchronously update statistics (non-critical, failure doesn't affect main flow)
	if s.statisticsService != nil {
		go s.statisticsService.UpdateStatisticsOnTaskStatusChange(
			context.Background(), task.Endpoint, oldStatus, task.Status)
	}

	logger.InfoCtx(ctx, "✅ orphaned task re-queued successfully, task_id: %s, endpoint: %s, status: PENDING",
		task.TaskID, task.Endpoint)
}

// CleanupTimedOutTasks checks for tasks that have exceeded their execution timeout and fails them
// This is a fallback mechanism for tasks stuck in IN_PROGRESS state when worker is unresponsive but still heartbeating
func (s *TaskService) CleanupTimedOutTasks(ctx context.Context) error {
	// Get global default timeout from config
	defaultTimeout := time.Duration(config.GlobalConfig.Queue.TaskTimeout) * time.Second

	logger.DebugCtx(ctx, "starting task timeout cleanup, default_timeout: %v", defaultTimeout)

	// Get all in-progress task IDs
	taskIDs, err := s.taskRepo.GetInProgressTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	if len(taskIDs) == 0 {
		logger.DebugCtx(ctx, "no in-progress tasks to check")
		return nil
	}

	logger.DebugCtx(ctx, "checking %d in-progress tasks for timeout", len(taskIDs))

	// OPTIMIZATION: Batch fetch all endpoint metadata upfront to avoid N+1 queries
	endpointTimeouts := make(map[string]time.Duration)
	if s.endpointService != nil {
		endpoints, err := s.endpointService.ListEndpoints(ctx)
		if err == nil {
			for _, ep := range endpoints {
				if ep.TaskTimeout > 0 {
					endpointTimeouts[ep.Name] = time.Duration(ep.TaskTimeout) * time.Second
				}
			}
		}
	}

	now := time.Now()
	timedOutCount := 0

	// TODO: Further optimization - batch fetch all tasks using pipeline/MGET
	for _, taskID := range taskIDs {
		mysqlTask, err := s.taskRepo.Get(ctx, taskID)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to get task during timeout check, task_id: %s, error: %v", taskID, err)
			continue
		}
		if mysqlTask == nil {
			continue
		}

		// Skip if task is no longer in progress (race condition)
		if mysqlTask.Status != string(model.TaskStatusInProgress) {
			continue
		}

		// Skip if task hasn't started yet (shouldn't happen, but be defensive)
		if mysqlTask.StartedAt == nil {
			continue
		}

		// Calculate how long the task has been running
		runningDuration := now.Sub(*mysqlTask.StartedAt)

		// Get timeout from pre-fetched endpoint metadata
		timeout := defaultTimeout
		if endpointTimeout, exists := endpointTimeouts[mysqlTask.Endpoint]; exists {
			timeout = endpointTimeout
			logger.DebugCtx(ctx, "using endpoint-specific timeout, task_id: %s, endpoint: %s, timeout: %v",
				taskID, mysqlTask.Endpoint, timeout)
		}

		// Check if task has timed out
		if runningDuration > timeout {
			logger.WarnCtx(ctx, "task execution timeout detected, task_id: %s, endpoint: %s, running_duration: %v, timeout: %v",
				taskID, mysqlTask.Endpoint, runningDuration, timeout)

			// Mark task as failed with timeout error
			oldStatus := mysqlTask.Status
			endpoint := mysqlTask.Endpoint
			mysqlTask.Status = string(model.TaskStatusFailed)
			mysqlTask.Error = fmt.Sprintf("Task execution timeout after %v (limit: %v)", runningDuration.Round(time.Second), timeout)
			completedAt := time.Now()
			mysqlTask.CompletedAt = &completedAt
			mysqlTask.UpdatedAt = completedAt

			// Record TASK_TIMEOUT event
			s.recordTaskTimeout(ctx, mysqlTask)

			if err := s.taskRepo.UpdateFields(ctx, mysqlTask.TaskID, map[string]interface{}{
				"status":       mysqlTask.Status,
				"error":        mysqlTask.Error,
				"completed_at": mysqlTask.CompletedAt,
				"updated_at":   mysqlTask.UpdatedAt,
			}); err != nil {
				logger.ErrorCtx(ctx, "failed to update timed-out task, task_id: %s, error: %v", taskID, err)
				continue
			}

			// Asynchronously update statistics (IN_PROGRESS -> FAILED)
			if s.statisticsService != nil {
				go s.statisticsService.UpdateStatisticsOnTaskStatusChange(context.Background(), endpoint, oldStatus, mysqlTask.Status)
			}

			// worker_id already cleared in MySQL (status = FAILED), no Redis operations needed
			timedOutCount++
			logger.InfoCtx(ctx, "task marked as failed due to timeout, task_id: %s, endpoint: %s, duration: %v",
				taskID, mysqlTask.Endpoint, runningDuration.Round(time.Second))
		}
	}

	if timedOutCount > 0 {
		logger.InfoCtx(ctx, "task timeout cleanup completed, timed_out: %d, total_checked: %d", timedOutCount, len(taskIDs))
	}

	return nil
}

// GetTaskExecutionHistory gets task execution history (task.extend field)
func (s *TaskService) GetTaskExecutionHistory(ctx context.Context, taskID string) ([]mysqlModel.ExecutionRecord, error) {
	mysqlTask, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return mysqlTask.GetExecutionHistory(), nil
}

// GetTaskEvents gets all task events (task_events table)
func (s *TaskService) GetTaskEvents(ctx context.Context, taskID string) ([]*mysql.TaskEvent, error) {
	return s.taskEventRepo.GetTaskEvents(ctx, taskID)
}

// GetTaskTimeline gets task timeline (simplified event list)
func (s *TaskService) GetTaskTimeline(ctx context.Context, taskID string) ([]*mysql.TaskEvent, error) {
	return s.taskEventRepo.GetTaskTimeline(ctx, taskID)
}
