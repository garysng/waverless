package monitoring

import (
	"context"
	"time"

	"waverless/pkg/logger"
	"waverless/pkg/store/mysql"
	"waverless/pkg/store/mysql/model"
)

// Aggregator handles monitoring data aggregation
type Aggregator struct {
	repo *mysql.MonitoringRepository
}

// NewAggregator creates a new aggregator
func NewAggregator(repo *mysql.MonitoringRepository) *Aggregator {
	return &Aggregator{repo: repo}
}

// AggregateMinuteStats aggregates statistics for the last minute
func (a *Aggregator) AggregateMinuteStats(ctx context.Context) error {
	now := time.Now().Truncate(time.Minute)
	from, to := now.Add(-time.Minute), now

	endpoints := a.getAllEndpoints(ctx, from, to)
	for endpoint := range endpoints {
		stat, err := a.repo.AggregateMinuteStats(ctx, endpoint, from, to)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to aggregate minute stats for %s: %v", endpoint, err)
			continue
		}
		a.repo.UpsertMinuteStat(ctx, stat)
	}

	a.repo.CleanupOldMinuteStats(ctx, now.Add(-12*time.Hour))
	return nil
}

// AggregateHourlyStats aggregates statistics for the last hour
func (a *Aggregator) AggregateHourlyStats(ctx context.Context) error {
	now := time.Now().Truncate(time.Hour)
	statHour := now.Add(-time.Hour)

	endpoints, _ := a.repo.GetAllEndpoints(ctx)
	for _, endpoint := range endpoints {
		stat, err := a.repo.AggregateHourlyStats(ctx, endpoint, statHour)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to aggregate hourly stats for %s: %v", endpoint, err)
			continue
		}
		a.repo.UpsertHourlyStat(ctx, stat)
	}

	a.repo.CleanupOldHourlyStats(ctx, now.AddDate(0, 0, -30))
	return nil
}

// AggregateDailyStats aggregates statistics for yesterday
func (a *Aggregator) AggregateDailyStats(ctx context.Context) error {
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())

	endpoints, _ := a.repo.GetAllEndpoints(ctx)
	for _, endpoint := range endpoints {
		stat, err := a.repo.AggregateDailyStats(ctx, endpoint, yesterday)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to aggregate daily stats for %s: %v", endpoint, err)
			continue
		}
		a.repo.UpsertDailyStat(ctx, stat)
	}

	a.repo.CleanupOldDailyStats(ctx, now.AddDate(0, 0, -90))
	return nil
}

// GetRealtimeMetrics returns real-time metrics for an endpoint
func (a *Aggregator) GetRealtimeMetrics(ctx context.Context, endpoint string) (*RealtimeMetrics, error) {
	m, err := a.repo.GetRealtimeMetrics(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return &RealtimeMetrics{
		Endpoint: m.Endpoint,
		Workers:  WorkerMetrics{Total: m.Workers.Total, Active: m.Workers.Active, Idle: m.Workers.Idle},
		Tasks:    TaskMetrics{InQueue: m.Tasks.InQueue, Running: m.Tasks.Running, CompletedLastMinute: m.Tasks.CompletedLastMinute},
		Performance: PerfMetrics{
			AvgQueueWaitMs:     m.Performance.AvgQueueWaitMs,
			AvgExecutionMs:     m.Performance.AvgExecutionMs,
			AvgTotalDurationMs: m.Performance.AvgTotalDurationMs,
		},
	}, nil
}

// GetMinuteStats returns minute-level statistics
func (a *Aggregator) GetMinuteStats(ctx context.Context, endpoint string, from, to time.Time) ([]*MinuteStatResponse, error) {
	stats, err := a.repo.GetMinuteStats(ctx, endpoint, from, to)
	if err != nil {
		return nil, err
	}
	return convertMinuteStats(stats), nil
}

// GetHourlyStats returns hourly statistics
func (a *Aggregator) GetHourlyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*HourlyStatResponse, error) {
	stats, err := a.repo.GetHourlyStats(ctx, endpoint, from, to)
	if err != nil {
		return nil, err
	}
	return convertHourlyStats(stats), nil
}

// GetDailyStats returns daily statistics
func (a *Aggregator) GetDailyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*DailyStatResponse, error) {
	stats, err := a.repo.GetDailyStats(ctx, endpoint, from, to)
	if err != nil {
		return nil, err
	}
	return convertDailyStats(stats), nil
}

func (a *Aggregator) getAllEndpoints(ctx context.Context, from, to time.Time) map[string]bool {
	endpoints, _ := a.repo.GetDistinctEndpoints(ctx, from, to)
	allEndpoints, _ := a.repo.GetAllEndpoints(ctx)
	set := make(map[string]bool)
	for _, ep := range endpoints {
		set[ep] = true
	}
	for _, ep := range allEndpoints {
		set[ep] = true
	}
	return set
}

func convertMinuteStats(stats []*model.EndpointMinuteStat) []*MinuteStatResponse {
	result := make([]*MinuteStatResponse, len(stats))
	for i, s := range stats {
		result[i] = &MinuteStatResponse{
			Timestamp:            s.StatMinute,
			ActiveWorkers:        s.ActiveWorkers,
			IdleWorkers:          s.IdleWorkers,
			AvgWorkerUtilization: s.AvgWorkerUtilization,
			TasksSubmitted:       s.TasksSubmitted,
			TasksCompleted:       s.TasksCompleted,
			TasksFailed:          s.TasksFailed,
			TasksTimeout:         s.TasksTimeout,
			TasksRetried:         s.TasksRetried,
			AvgQueueWaitMs:       s.AvgQueueWaitMs,
			AvgExecutionMs:       s.AvgExecutionMs,
			P50ExecutionMs:       s.P50ExecutionMs,
			P95ExecutionMs:       s.P95ExecutionMs,
			AvgIdleDurationSec:   s.AvgIdleDurationSec,
			MaxIdleDurationSec:   s.MaxIdleDurationSec,
			TotalIdleTimeSec:     s.TotalIdleTimeSec,
			IdleCount:            s.IdleCount,
			WorkersCreated:       s.WorkersCreated,
			WorkersTerminated:    s.WorkersTerminated,
			ColdStarts:           s.ColdStarts,
			AvgColdStartMs:       s.AvgColdStartMs,
		}
	}
	return result
}

func convertHourlyStats(stats []*model.EndpointHourlyStat) []*HourlyStatResponse {
	result := make([]*HourlyStatResponse, len(stats))
	for i, s := range stats {
		result[i] = &HourlyStatResponse{
			Timestamp:            s.StatHour,
			ActiveWorkers:        s.ActiveWorkers,
			IdleWorkers:          s.IdleWorkers,
			AvgWorkerUtilization: s.AvgWorkerUtilization,
			TasksSubmitted:       s.TasksSubmitted,
			TasksCompleted:       s.TasksCompleted,
			TasksFailed:          s.TasksFailed,
			TasksTimeout:         s.TasksTimeout,
			TasksRetried:         s.TasksRetried,
			AvgQueueWaitMs:       s.AvgQueueWaitMs,
			AvgExecutionMs:       s.AvgExecutionMs,
			P50ExecutionMs:       s.P50ExecutionMs,
			P95ExecutionMs:       s.P95ExecutionMs,
			AvgIdleDurationSec:   s.AvgIdleDurationSec,
			MaxIdleDurationSec:   s.MaxIdleDurationSec,
			TotalIdleTimeSec:     s.TotalIdleTimeSec,
			IdleCount:            s.IdleCount,
			WorkersCreated:       s.WorkersCreated,
			WorkersTerminated:    s.WorkersTerminated,
			ColdStarts:           s.ColdStarts,
			AvgColdStartMs:       s.AvgColdStartMs,
		}
	}
	return result
}

func convertDailyStats(stats []*model.EndpointDailyStat) []*DailyStatResponse {
	result := make([]*DailyStatResponse, len(stats))
	for i, s := range stats {
		result[i] = &DailyStatResponse{
			Date:                 s.StatDate,
			ActiveWorkers:        s.ActiveWorkers,
			IdleWorkers:          s.IdleWorkers,
			AvgWorkerUtilization: s.AvgWorkerUtilization,
			TasksSubmitted:       s.TasksSubmitted,
			TasksCompleted:       s.TasksCompleted,
			TasksFailed:          s.TasksFailed,
			TasksTimeout:         s.TasksTimeout,
			TasksRetried:         s.TasksRetried,
			AvgQueueWaitMs:       s.AvgQueueWaitMs,
			AvgExecutionMs:       s.AvgExecutionMs,
			P50ExecutionMs:       s.P50ExecutionMs,
			P95ExecutionMs:       s.P95ExecutionMs,
			AvgIdleDurationSec:   s.AvgIdleDurationSec,
			MaxIdleDurationSec:   s.MaxIdleDurationSec,
			TotalIdleTimeSec:     s.TotalIdleTimeSec,
			IdleCount:            s.IdleCount,
			WorkersCreated:       s.WorkersCreated,
			WorkersTerminated:    s.WorkersTerminated,
			ColdStarts:           s.ColdStarts,
			AvgColdStartMs:       s.AvgColdStartMs,
		}
	}
	return result
}
