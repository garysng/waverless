package service

import (
	"context"
	"time"

	"waverless/pkg/monitoring"
	"waverless/pkg/store/mysql"
)

// MonitoringService wraps monitoring.Aggregator for backward compatibility
type MonitoringService struct {
	agg *monitoring.Aggregator
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(monitoringRepo *mysql.MonitoringRepository) *MonitoringService {
	return &MonitoringService{agg: monitoring.NewAggregator(monitoringRepo)}
}

func (s *MonitoringService) AggregateMinuteStats(ctx context.Context) error {
	return s.agg.AggregateMinuteStats(ctx)
}

func (s *MonitoringService) AggregateHourlyStats(ctx context.Context) error {
	return s.agg.AggregateHourlyStats(ctx)
}

func (s *MonitoringService) AggregateDailyStats(ctx context.Context) error {
	return s.agg.AggregateDailyStats(ctx)
}

func (s *MonitoringService) GetRealtimeMetrics(ctx context.Context, endpoint string) (*monitoring.RealtimeMetrics, error) {
	return s.agg.GetRealtimeMetrics(ctx, endpoint)
}

func (s *MonitoringService) GetMinuteStats(ctx context.Context, endpoint string, from, to time.Time) ([]*monitoring.MinuteStatResponse, error) {
	return s.agg.GetMinuteStats(ctx, endpoint, from, to)
}

func (s *MonitoringService) GetHourlyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*monitoring.HourlyStatResponse, error) {
	return s.agg.GetHourlyStats(ctx, endpoint, from, to)
}

func (s *MonitoringService) GetDailyStats(ctx context.Context, endpoint string, from, to time.Time) ([]*monitoring.DailyStatResponse, error) {
	return s.agg.GetDailyStats(ctx, endpoint, from, to)
}
