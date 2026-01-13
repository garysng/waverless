package monitoring

import (
	"context"

	"waverless/pkg/store/mysql"
)

// Collector is deprecated - worker metrics now come from worker_events table
// Kept for backward compatibility but does nothing
type Collector struct {
	repo *mysql.MonitoringRepository
}

// NewCollector creates a new collector (no-op)
func NewCollector(repo *mysql.MonitoringRepository, workerRepo *mysql.WorkerRepository, taskRepo *mysql.TaskRepository) *Collector {
	return &Collector{repo: repo}
}

// CollectSnapshots is deprecated - no longer needed
func (c *Collector) CollectSnapshots(ctx context.Context) error {
	return nil
}
