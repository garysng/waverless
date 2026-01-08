package mysql

import (
	"context"
	"time"

	"waverless/pkg/store/mysql/model"

	"gorm.io/gorm/clause"
)

// Helper functions for type conversion from map
func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

// MonitoringRepository handles monitoring statistics database operations
type MonitoringRepository struct {
	ds *Datastore
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(ds *Datastore) *MonitoringRepository {
	return &MonitoringRepository{ds: ds}
}

// SaveResourceSnapshot saves a worker resource snapshot
func (r *MonitoringRepository) SaveResourceSnapshot(ctx context.Context, snapshot *model.WorkerResourceSnapshot) error {
	return r.ds.DB(ctx).Create(snapshot).Error
}

// CleanupOldSnapshots removes snapshots older than the given time
func (r *MonitoringRepository) CleanupOldSnapshots(ctx context.Context, before time.Time) (int64, error) {
	result := r.ds.DB(ctx).Where("snapshot_at < ?", before).Delete(&model.WorkerResourceSnapshot{})
	return result.RowsAffected, result.Error
}

// UpsertMinuteStat creates or updates minute-level statistics
func (r *MonitoringRepository) UpsertMinuteStat(ctx context.Context, stat *model.EndpointMinuteStat) error {
	return r.ds.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}, {Name: "stat_minute"}},
		UpdateAll: true,
	}).Create(stat).Error
}

// GetMinuteStats retrieves minute-level statistics for a time range
func (r *MonitoringRepository) GetMinuteStats(ctx context.Context, endpoint string, from, to time.Time) ([]*model.EndpointMinuteStat, error) {
	var stats []*model.EndpointMinuteStat
	err := r.ds.DB(ctx).Where("endpoint = ? AND stat_minute >= ? AND stat_minute < ?", endpoint, from, to).
		Order("stat_minute ASC").Find(&stats).Error
	return stats, err
}

// UpsertHourlyStat creates or updates hourly statistics
func (r *MonitoringRepository) UpsertHourlyStat(ctx context.Context, stat *model.EndpointHourlyStat) error {
	return r.ds.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}, {Name: "stat_hour"}},
		UpdateAll: true,
	}).Create(stat).Error
}

// GetHourlyStats retrieves hourly statistics for a time range
func (r *MonitoringRepository) GetHourlyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*model.EndpointHourlyStat, error) {
	var stats []*model.EndpointHourlyStat
	err := r.ds.DB(ctx).Where("endpoint = ? AND stat_hour >= ? AND stat_hour < ?", endpoint, from, to).
		Order("stat_hour ASC").Find(&stats).Error
	return stats, err
}

// UpsertDailyStat creates or updates daily statistics
func (r *MonitoringRepository) UpsertDailyStat(ctx context.Context, stat *model.EndpointDailyStat) error {
	return r.ds.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}, {Name: "stat_date"}},
		UpdateAll: true,
	}).Create(stat).Error
}

// GetDailyStats retrieves daily statistics for a date range
func (r *MonitoringRepository) GetDailyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*model.EndpointDailyStat, error) {
	var stats []*model.EndpointDailyStat
	err := r.ds.DB(ctx).Where("endpoint = ? AND stat_date >= ? AND stat_date < ?", endpoint, from, to).
		Order("stat_date ASC").Find(&stats).Error
	return stats, err
}

// CleanupOldMinuteStats removes minute stats older than retention period
func (r *MonitoringRepository) CleanupOldMinuteStats(ctx context.Context, before time.Time) (int64, error) {
	result := r.ds.DB(ctx).Where("stat_minute < ?", before).Delete(&model.EndpointMinuteStat{})
	return result.RowsAffected, result.Error
}

// CleanupOldHourlyStats removes hourly stats older than retention period
func (r *MonitoringRepository) CleanupOldHourlyStats(ctx context.Context, before time.Time) (int64, error) {
	result := r.ds.DB(ctx).Where("stat_hour < ?", before).Delete(&model.EndpointHourlyStat{})
	return result.RowsAffected, result.Error
}

// CleanupOldDailyStats removes daily stats older than retention period
func (r *MonitoringRepository) CleanupOldDailyStats(ctx context.Context, before time.Time) (int64, error) {
	result := r.ds.DB(ctx).Where("stat_date < ?", before).Delete(&model.EndpointDailyStat{})
	return result.RowsAffected, result.Error
}

// AggregateMinuteStats aggregates task events into minute-level statistics
func (r *MonitoringRepository) AggregateMinuteStats(ctx context.Context, endpoint string, from, to time.Time) (*model.EndpointMinuteStat, error) {
	stat := &model.EndpointMinuteStat{Endpoint: endpoint, StatMinute: from}

	// Task stats from task_events
	var taskStats struct {
		TasksSubmitted int     `gorm:"column:tasks_submitted"`
		TasksCompleted int     `gorm:"column:tasks_completed"`
		TasksFailed    int     `gorm:"column:tasks_failed"`
		TasksTimeout   int     `gorm:"column:tasks_timeout"`
		AvgQueueWaitMs float64 `gorm:"column:avg_queue_wait_ms"`
		AvgExecutionMs float64 `gorm:"column:avg_execution_ms"`
	}
	r.ds.DB(ctx).Raw(`
		SELECT
			COUNT(CASE WHEN event_type = 'TASK_CREATED' THEN 1 END) as tasks_submitted,
			COUNT(CASE WHEN event_type = 'TASK_COMPLETED' THEN 1 END) as tasks_completed,
			COUNT(CASE WHEN event_type = 'TASK_FAILED' THEN 1 END) as tasks_failed,
			COUNT(CASE WHEN event_type = 'TASK_TIMEOUT' THEN 1 END) as tasks_timeout,
			COALESCE(AVG(CASE WHEN event_type IN ('TASK_COMPLETED','TASK_FAILED','TASK_TIMEOUT') THEN queue_wait_ms END), 0) as avg_queue_wait_ms,
			COALESCE(AVG(CASE WHEN event_type IN ('TASK_COMPLETED','TASK_FAILED','TASK_TIMEOUT') THEN execution_duration_ms END), 0) as avg_execution_ms
		FROM task_events WHERE endpoint = ? AND event_time >= ? AND event_time < ?
	`, endpoint, from, to).Scan(&taskStats)
	stat.TasksSubmitted = taskStats.TasksSubmitted
	stat.TasksCompleted = taskStats.TasksCompleted
	stat.TasksFailed = taskStats.TasksFailed
	stat.TasksTimeout = taskStats.TasksTimeout
	stat.AvgQueueWaitMs = taskStats.AvgQueueWaitMs
	stat.AvgExecutionMs = taskStats.AvgExecutionMs

	// P95 execution time (simple approximation: use max as P95 for small datasets)
	var p95 float64
	r.ds.DB(ctx).Raw(`
		SELECT COALESCE(MAX(execution_duration_ms), 0) FROM task_events 
		WHERE endpoint = ? AND event_time >= ? AND event_time < ? 
		AND event_type IN ('TASK_COMPLETED','TASK_FAILED','TASK_TIMEOUT') AND execution_duration_ms IS NOT NULL
	`, endpoint, from, to).Scan(&p95)
	stat.P95ExecutionMs = p95

	// Worker stats (no gpu_utilization in workers table)
	var workerStats struct {
		ActiveWorkers int `gorm:"column:active_workers"`
		IdleWorkers   int `gorm:"column:idle_workers"`
	}
	r.ds.DB(ctx).Raw(`
		SELECT
			COUNT(CASE WHEN status IN ('ONLINE', 'BUSY') THEN 1 END) as active_workers,
			COUNT(CASE WHEN status = 'ONLINE' AND current_jobs = 0 THEN 1 END) as idle_workers
		FROM workers WHERE endpoint = ? AND last_heartbeat >= ?
	`, endpoint, from).Scan(&workerStats)
	stat.ActiveWorkers = workerStats.ActiveWorkers
	stat.IdleWorkers = workerStats.IdleWorkers

	// GPU utilization from snapshots
	var avgGPU float64
	r.ds.DB(ctx).Raw(`
		SELECT COALESCE(AVG(gpu_utilization), 0) FROM worker_resource_snapshots s
		JOIN workers w ON s.worker_id = w.worker_id
		WHERE w.endpoint = ? AND s.snapshot_at >= ? AND s.snapshot_at < ?
	`, endpoint, from, to).Scan(&avgGPU)
	stat.AvgGPUUtilization = avgGPU

	// Cold start stats
	var coldStats struct {
		ColdStarts     int     `gorm:"column:cold_starts"`
		AvgColdStartMs float64 `gorm:"column:avg_cold_start_ms"`
	}
	r.ds.DB(ctx).Raw(`
		SELECT COUNT(*) as cold_starts, COALESCE(AVG(cold_start_duration_ms), 0) as avg_cold_start_ms
		FROM workers WHERE endpoint = ? AND cold_start_duration_ms IS NOT NULL AND pod_started_at >= ? AND pod_started_at < ?
	`, endpoint, from, to).Scan(&coldStats)
	stat.ColdStarts = coldStats.ColdStarts
	stat.AvgColdStartMs = coldStats.AvgColdStartMs

	return stat, nil
}

// AggregateHourlyStats aggregates minute stats into hourly statistics
func (r *MonitoringRepository) AggregateHourlyStats(ctx context.Context, endpoint string, statHour time.Time) (*model.EndpointHourlyStat, error) {
	stat := &model.EndpointHourlyStat{Endpoint: endpoint, StatHour: statHour}
	from, to := statHour, statHour.Add(time.Hour)

	// Check if there's any data first
	var count int64
	r.ds.DB(ctx).Model(&model.EndpointMinuteStat{}).Where("endpoint = ? AND stat_minute >= ? AND stat_minute < ?", endpoint, from, to).Count(&count)
	if count == 0 {
		return stat, nil
	}

	// Use map to avoid field misalignment issues with Raw().Scan()
	var result map[string]interface{}
	r.ds.DB(ctx).Raw(`
		SELECT
			COALESCE(AVG(active_workers), 0) as active_workers,
			COALESCE(AVG(idle_workers), 0) as idle_workers,
			COALESCE(SUM(tasks_submitted), 0) as tasks_submitted,
			COALESCE(SUM(tasks_completed), 0) as tasks_completed,
			COALESCE(SUM(tasks_failed), 0) as tasks_failed,
			COALESCE(SUM(tasks_timeout), 0) as tasks_timeout,
			COALESCE(AVG(avg_queue_wait_ms), 0) as avg_queue_wait_ms,
			COALESCE(AVG(avg_execution_ms), 0) as avg_execution_ms,
			COALESCE(MAX(p95_execution_ms), 0) as p95_execution_ms,
			COALESCE(AVG(avg_gpu_utilization), 0) as avg_gpu_utilization,
			COALESCE(SUM(cold_starts), 0) as cold_starts,
			COALESCE(AVG(avg_cold_start_ms), 0) as avg_cold_start_ms
		FROM endpoint_minute_stats WHERE endpoint = ? AND stat_minute >= ? AND stat_minute < ?
	`, endpoint, from, to).Scan(&result)

	if result != nil {
		stat.ActiveWorkers = toInt(result["active_workers"])
		stat.IdleWorkers = toInt(result["idle_workers"])
		stat.TasksSubmitted = toInt(result["tasks_submitted"])
		stat.TasksCompleted = toInt(result["tasks_completed"])
		stat.TasksFailed = toInt(result["tasks_failed"])
		stat.TasksTimeout = toInt(result["tasks_timeout"])
		stat.AvgQueueWaitMs = toFloat(result["avg_queue_wait_ms"])
		stat.AvgExecutionMs = toFloat(result["avg_execution_ms"])
		stat.P95ExecutionMs = toFloat(result["p95_execution_ms"])
		stat.AvgGPUUtilization = toFloat(result["avg_gpu_utilization"])
		stat.ColdStarts = toInt(result["cold_starts"])
		stat.AvgColdStartMs = toFloat(result["avg_cold_start_ms"])
	}

	return stat, nil
}

// AggregateDailyStats aggregates hourly stats into daily statistics
func (r *MonitoringRepository) AggregateDailyStats(ctx context.Context, endpoint string, statDate time.Time) (*model.EndpointDailyStat, error) {
	stat := &model.EndpointDailyStat{Endpoint: endpoint, StatDate: statDate}
	from, to := statDate, statDate.AddDate(0, 0, 1)

	// Check if there's any data first
	var count int64
	r.ds.DB(ctx).Model(&model.EndpointHourlyStat{}).Where("endpoint = ? AND stat_hour >= ? AND stat_hour < ?", endpoint, from, to).Count(&count)
	if count == 0 {
		return stat, nil
	}

	var result map[string]interface{}
	r.ds.DB(ctx).Raw(`
		SELECT
			COALESCE(AVG(active_workers), 0) as active_workers,
			COALESCE(AVG(idle_workers), 0) as idle_workers,
			COALESCE(SUM(tasks_submitted), 0) as tasks_submitted,
			COALESCE(SUM(tasks_completed), 0) as tasks_completed,
			COALESCE(SUM(tasks_failed), 0) as tasks_failed,
			COALESCE(SUM(tasks_timeout), 0) as tasks_timeout,
			COALESCE(AVG(avg_queue_wait_ms), 0) as avg_queue_wait_ms,
			COALESCE(AVG(avg_execution_ms), 0) as avg_execution_ms,
			COALESCE(MAX(p95_execution_ms), 0) as p95_execution_ms,
			COALESCE(AVG(avg_gpu_utilization), 0) as avg_gpu_utilization,
			COALESCE(SUM(cold_starts), 0) as cold_starts,
			COALESCE(AVG(avg_cold_start_ms), 0) as avg_cold_start_ms
		FROM endpoint_hourly_stats WHERE endpoint = ? AND stat_hour >= ? AND stat_hour < ?
	`, endpoint, from, to).Scan(&result)

	if result != nil {
		stat.ActiveWorkers = toInt(result["active_workers"])
		stat.IdleWorkers = toInt(result["idle_workers"])
		stat.TasksSubmitted = toInt(result["tasks_submitted"])
		stat.TasksCompleted = toInt(result["tasks_completed"])
		stat.TasksFailed = toInt(result["tasks_failed"])
		stat.TasksTimeout = toInt(result["tasks_timeout"])
		stat.AvgQueueWaitMs = toFloat(result["avg_queue_wait_ms"])
		stat.AvgExecutionMs = toFloat(result["avg_execution_ms"])
		stat.P95ExecutionMs = toFloat(result["p95_execution_ms"])
		stat.AvgGPUUtilization = toFloat(result["avg_gpu_utilization"])
		stat.ColdStarts = toInt(result["cold_starts"])
		stat.AvgColdStartMs = toFloat(result["avg_cold_start_ms"])
	}

	return stat, nil
}

// GetDistinctEndpoints returns all distinct endpoints from task_events
func (r *MonitoringRepository) GetDistinctEndpoints(ctx context.Context, from, to time.Time) ([]string, error) {
	var endpoints []string
	err := r.ds.DB(ctx).Model(&model.TaskEvent{}).
		Where("event_time >= ? AND event_time < ?", from, to).
		Distinct().Pluck("endpoint", &endpoints).Error
	return endpoints, err
}

// GetAllEndpoints returns all non-deleted endpoints from endpoints table
func (r *MonitoringRepository) GetAllEndpoints(ctx context.Context) ([]string, error) {
	var endpoints []string
	err := r.ds.DB(ctx).Model(&model.Endpoint{}).Where("status != ?", "deleted").Pluck("endpoint", &endpoints).Error
	return endpoints, err
}

// GetRealtimeMetrics returns real-time metrics for an endpoint
func (r *MonitoringRepository) GetRealtimeMetrics(ctx context.Context, endpoint string) (*RealtimeMetrics, error) {
	metrics := &RealtimeMetrics{Endpoint: endpoint}
	threshold := time.Now().Add(-60 * time.Second)

	r.ds.DB(ctx).Raw(`
		SELECT COUNT(*) as total,
			COUNT(CASE WHEN status IN ('ONLINE', 'BUSY') THEN 1 END) as active,
			COUNT(CASE WHEN status = 'ONLINE' AND current_jobs = 0 THEN 1 END) as idle
		FROM workers WHERE endpoint = ? AND last_heartbeat > ?
	`, endpoint, threshold).Scan(&metrics.Workers)

	lastMinute := time.Now().Add(-time.Minute)
	r.ds.DB(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM tasks WHERE endpoint = ? AND status = 'PENDING') as in_queue,
			(SELECT COUNT(*) FROM tasks WHERE endpoint = ? AND status = 'IN_PROGRESS') as running,
			(SELECT COUNT(*) FROM task_events WHERE endpoint = ? AND event_type = 'TASK_COMPLETED' AND event_time >= ?) as completed_last_minute
	`, endpoint, endpoint, endpoint, lastMinute).Scan(&metrics.Tasks)

	last5Min := time.Now().Add(-5 * time.Minute)
	r.ds.DB(ctx).Raw(`
		SELECT COALESCE(AVG(queue_wait_ms), 0) as avg_queue_wait_ms,
			COALESCE(AVG(execution_duration_ms), 0) as avg_execution_ms,
			COALESCE(AVG(total_duration_ms), 0) as avg_total_duration_ms
		FROM task_events WHERE endpoint = ? AND event_type IN ('TASK_COMPLETED','TASK_FAILED') AND event_time >= ?
	`, endpoint, last5Min).Scan(&metrics.Performance)

	return metrics, nil
}

// RealtimeMetrics represents real-time metrics for an endpoint
type RealtimeMetrics struct {
	Endpoint    string `json:"endpoint"`
	Workers     struct {
		Total  int `json:"total"`
		Active int `json:"active"`
		Idle   int `json:"idle"`
	} `json:"workers"`
	Tasks struct {
		InQueue             int64 `json:"in_queue"`
		Running             int64 `json:"running"`
		CompletedLastMinute int   `json:"completed_last_minute"`
	} `json:"tasks"`
	Performance struct {
		AvgQueueWaitMs     float64 `json:"avg_queue_wait_ms"`
		AvgExecutionMs     float64 `json:"avg_execution_ms"`
		AvgTotalDurationMs float64 `json:"avg_total_duration_ms"`
	} `json:"performance"`
}



