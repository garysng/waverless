package monitoring

import "time"

// RealtimeMetrics represents real-time metrics for an endpoint
type RealtimeMetrics struct {
	Endpoint    string        `json:"endpoint"`
	Workers     WorkerMetrics `json:"workers"`
	Tasks       TaskMetrics   `json:"tasks"`
	Performance PerfMetrics   `json:"performance"`
}

type WorkerMetrics struct {
	Total  int `json:"total"`
	Active int `json:"active"`
	Idle   int `json:"idle"`
}

type TaskMetrics struct {
	InQueue             int64 `json:"in_queue"`
	Running             int64 `json:"running"`
	CompletedLastMinute int   `json:"completed_last_minute"`
}

type PerfMetrics struct {
	AvgQueueWaitMs     float64 `json:"avg_queue_wait_ms"`
	AvgExecutionMs     float64 `json:"avg_execution_ms"`
	AvgTotalDurationMs float64 `json:"avg_total_duration_ms"`
}

// MinuteStatResponse represents minute-level stats API response
type MinuteStatResponse struct {
	Timestamp            time.Time `json:"timestamp"`
	ActiveWorkers        int       `json:"active_workers"`
	IdleWorkers          int       `json:"idle_workers"`
	AvgWorkerUtilization float64   `json:"avg_worker_utilization"`
	TasksSubmitted       int       `json:"tasks_submitted"`
	TasksCompleted       int       `json:"tasks_completed"`
	TasksFailed          int       `json:"tasks_failed"`
	TasksTimeout         int       `json:"tasks_timeout"`
	TasksRetried         int       `json:"tasks_retried"`
	AvgQueueWaitMs       float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs       float64   `json:"avg_execution_ms"`
	P50ExecutionMs       float64   `json:"p50_execution_ms"`
	P95ExecutionMs       float64   `json:"p95_execution_ms"`
	AvgGPUUtilization    float64   `json:"avg_gpu_utilization"`
	MaxGPUUtilization    float64   `json:"max_gpu_utilization"`
	AvgIdleDurationSec   float64   `json:"avg_idle_duration_sec"`
	MaxIdleDurationSec   int       `json:"max_idle_duration_sec"`
	TotalIdleTimeSec     int       `json:"total_idle_time_sec"`
	IdleCount            int       `json:"idle_count"`
	WorkersCreated       int       `json:"workers_created"`
	WorkersTerminated    int       `json:"workers_terminated"`
	ColdStarts           int       `json:"cold_starts"`
	AvgColdStartMs       float64   `json:"avg_cold_start_ms"`
	WebhookSuccess       int       `json:"webhook_success"`
	WebhookFailed        int       `json:"webhook_failed"`
}

// HourlyStatResponse represents hourly stats API response
type HourlyStatResponse struct {
	Timestamp            time.Time `json:"timestamp"`
	ActiveWorkers        int       `json:"active_workers"`
	IdleWorkers          int       `json:"idle_workers"`
	AvgWorkerUtilization float64   `json:"avg_worker_utilization"`
	TasksSubmitted       int       `json:"tasks_submitted"`
	TasksCompleted       int       `json:"tasks_completed"`
	TasksFailed          int       `json:"tasks_failed"`
	TasksTimeout         int       `json:"tasks_timeout"`
	TasksRetried         int       `json:"tasks_retried"`
	AvgQueueWaitMs       float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs       float64   `json:"avg_execution_ms"`
	P50ExecutionMs       float64   `json:"p50_execution_ms"`
	P95ExecutionMs       float64   `json:"p95_execution_ms"`
	AvgGPUUtilization    float64   `json:"avg_gpu_utilization"`
	MaxGPUUtilization    float64   `json:"max_gpu_utilization"`
	AvgIdleDurationSec   float64   `json:"avg_idle_duration_sec"`
	MaxIdleDurationSec   int       `json:"max_idle_duration_sec"`
	TotalIdleTimeSec     int64     `json:"total_idle_time_sec"`
	IdleCount            int       `json:"idle_count"`
	WorkersCreated       int       `json:"workers_created"`
	WorkersTerminated    int       `json:"workers_terminated"`
	ColdStarts           int       `json:"cold_starts"`
	AvgColdStartMs       float64   `json:"avg_cold_start_ms"`
	WebhookSuccess       int       `json:"webhook_success"`
	WebhookFailed        int       `json:"webhook_failed"`
}

// DailyStatResponse represents daily stats API response
type DailyStatResponse struct {
	Date                 time.Time `json:"date"`
	ActiveWorkers        int       `json:"active_workers"`
	IdleWorkers          int       `json:"idle_workers"`
	AvgWorkerUtilization float64   `json:"avg_worker_utilization"`
	TasksSubmitted       int       `json:"tasks_submitted"`
	TasksCompleted       int       `json:"tasks_completed"`
	TasksFailed          int       `json:"tasks_failed"`
	TasksTimeout         int       `json:"tasks_timeout"`
	TasksRetried         int       `json:"tasks_retried"`
	AvgQueueWaitMs       float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs       float64   `json:"avg_execution_ms"`
	P50ExecutionMs       float64   `json:"p50_execution_ms"`
	P95ExecutionMs       float64   `json:"p95_execution_ms"`
	AvgGPUUtilization    float64   `json:"avg_gpu_utilization"`
	MaxGPUUtilization    float64   `json:"max_gpu_utilization"`
	AvgIdleDurationSec   float64   `json:"avg_idle_duration_sec"`
	MaxIdleDurationSec   int       `json:"max_idle_duration_sec"`
	TotalIdleTimeSec     int64     `json:"total_idle_time_sec"`
	IdleCount            int       `json:"idle_count"`
	WorkersCreated       int       `json:"workers_created"`
	WorkersTerminated    int       `json:"workers_terminated"`
	ColdStarts           int       `json:"cold_starts"`
	AvgColdStartMs       float64   `json:"avg_cold_start_ms"`
	WebhookSuccess       int       `json:"webhook_success"`
	WebhookFailed        int       `json:"webhook_failed"`
}
