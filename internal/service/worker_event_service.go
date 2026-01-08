package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"waverless/pkg/store/mysql"
	"waverless/pkg/store/mysql/model"
)

// WorkerEventService records worker lifecycle events
type WorkerEventService struct {
	monitoringRepo *mysql.MonitoringRepository
}

// NewWorkerEventService creates a new worker event service
func NewWorkerEventService(monitoringRepo *mysql.MonitoringRepository) *WorkerEventService {
	return &WorkerEventService{monitoringRepo: monitoringRepo}
}

// RecordWorkerRegistered records when a worker becomes ready (first heartbeat)
func (s *WorkerEventService) RecordWorkerRegistered(ctx context.Context, workerID, endpoint, podName string, coldStartMs *int64) {
	event := &model.WorkerEvent{
		EventID:             uuid.New().String(),
		WorkerID:            workerID,
		Endpoint:            endpoint,
		EventType:           string(model.EventWorkerRegistered),
		EventTime:           time.Now(),
		ColdStartDurationMs: coldStartMs,
	}
	s.monitoringRepo.SaveWorkerEvent(ctx, event)
}

// RecordWorkerTaskPulled records when a worker pulls a task (with idle duration)
func (s *WorkerEventService) RecordWorkerTaskPulled(ctx context.Context, workerID, endpoint, podName, taskID string, idleDurationMs int64) {
	event := &model.WorkerEvent{
		EventID:        uuid.New().String(),
		WorkerID:       workerID,
		Endpoint:       endpoint,
		EventType:      string(model.EventWorkerTaskPulled),
		EventTime:      time.Now(),
		TaskID:         taskID,
		IdleDurationMs: &idleDurationMs,
	}
	s.monitoringRepo.SaveWorkerEvent(ctx, event)
}

// RecordWorkerOffline records when a worker goes offline
func (s *WorkerEventService) RecordWorkerOffline(ctx context.Context, workerID, endpoint, podName string) {
	event := &model.WorkerEvent{
		EventID:   uuid.New().String(),
		WorkerID:  workerID,
		Endpoint:  endpoint,
		EventType: string(model.EventWorkerOffline),
		EventTime: time.Now(),
	}
	s.monitoringRepo.SaveWorkerEvent(ctx, event)
}

// RecordWorkerStarted records when a pod starts
func (s *WorkerEventService) RecordWorkerStarted(ctx context.Context, workerID, endpoint string) {
	event := &model.WorkerEvent{
		EventID:   uuid.New().String(),
		WorkerID:  workerID,
		Endpoint:  endpoint,
		EventType: string(model.EventWorkerStarted),
		EventTime: time.Now(),
	}
	s.monitoringRepo.SaveWorkerEvent(ctx, event)
}

// RecordWorkerTaskCompleted records when a worker completes a task
func (s *WorkerEventService) RecordWorkerTaskCompleted(ctx context.Context, workerID, endpoint, taskID string, executionMs int64) {
	event := &model.WorkerEvent{
		EventID:   uuid.New().String(),
		WorkerID:  workerID,
		Endpoint:  endpoint,
		EventType: string(model.EventWorkerTaskCompleted),
		EventTime: time.Now(),
		TaskID:    taskID,
	}
	s.monitoringRepo.SaveWorkerEvent(ctx, event)
}
