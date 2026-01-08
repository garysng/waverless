package monitoring

import "time"

// RealtimeMetrics represents real-time metrics for an endpoint
type RealtimeMetrics struct {
	Endpoint    string          `json:"endpoint"`
	Workers     WorkerMetrics   `json:"workers"`
	Tasks       TaskMetrics     `json:"tasks"`
	Performance PerfMetrics     `json:"performance"`
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
	Timestamp      time.Time `json:"timestamp"`
	ActiveWorkers  int       `json:"active_workers"`
	IdleWorkers    int       `json:"idle_workers"`
	TasksSubmitted int       `json:"tasks_submitted"`
	TasksCompleted int       `json:"tasks_completed"`
	TasksFailed    int       `json:"tasks_failed"`
	TasksTimeout   int       `json:"tasks_timeout"`
	AvgQueueWaitMs float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs float64   `json:"avg_execution_ms"`
	P95ExecutionMs float64   `json:"p95_execution_ms"`
}

// HourlyStatResponse represents hourly stats API response
type HourlyStatResponse struct {
	Timestamp      time.Time `json:"timestamp"`
	ActiveWorkers  int       `json:"active_workers"`
	IdleWorkers    int       `json:"idle_workers"`
	TasksSubmitted int       `json:"tasks_submitted"`
	TasksCompleted int       `json:"tasks_completed"`
	TasksFailed    int       `json:"tasks_failed"`
	TasksTimeout   int       `json:"tasks_timeout"`
	AvgQueueWaitMs float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs float64   `json:"avg_execution_ms"`
	P95ExecutionMs float64   `json:"p95_execution_ms"`
	ColdStarts     int       `json:"cold_starts"`
	AvgColdStartMs float64   `json:"avg_cold_start_ms"`
}

// DailyStatResponse represents daily stats API response
type DailyStatResponse struct {
	Date           time.Time `json:"date"`
	ActiveWorkers  int       `json:"active_workers"`
	IdleWorkers    int       `json:"idle_workers"`
	TasksSubmitted int       `json:"tasks_submitted"`
	TasksCompleted int       `json:"tasks_completed"`
	TasksFailed    int       `json:"tasks_failed"`
	TasksTimeout   int       `json:"tasks_timeout"`
	AvgQueueWaitMs float64   `json:"avg_queue_wait_ms"`
	AvgExecutionMs float64   `json:"avg_execution_ms"`
	P95ExecutionMs float64   `json:"p95_execution_ms"`
	ColdStarts     int       `json:"cold_starts"`
	AvgColdStartMs float64   `json:"avg_cold_start_ms"`
}
