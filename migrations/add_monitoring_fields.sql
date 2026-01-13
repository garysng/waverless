-- Migration: Add missing monitoring fields
-- Date: 2026-01-08

-- 1. Add endpoint to worker_resource_snapshots (avoid JOIN during aggregation)
ALTER TABLE `worker_resource_snapshots` ADD COLUMN  `endpoint` VARCHAR(255) AFTER `worker_id`;
ALTER TABLE `worker_resource_snapshots` ADD INDEX  `idx_endpoint_snapshot` (`endpoint`, `snapshot_at`);

-- 2. Add missing fields to endpoint_minute_stats
ALTER TABLE `endpoint_minute_stats`
  ADD COLUMN  `tasks_retried` INT DEFAULT 0 AFTER `tasks_timeout`,
  ADD COLUMN  `webhook_success` INT DEFAULT 0 AFTER `avg_cold_start_ms`,
  ADD COLUMN  `webhook_failed` INT DEFAULT 0 AFTER `webhook_success`,
  ADD COLUMN  `avg_worker_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `idle_workers`,
  ADD COLUMN  `max_gpu_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `avg_gpu_utilization`,
  ADD COLUMN  `avg_idle_duration_sec` DECIMAL(10,2) DEFAULT 0 AFTER `max_gpu_utilization`,
  ADD COLUMN  `max_idle_duration_sec` INT DEFAULT 0 AFTER `avg_idle_duration_sec`,
  ADD COLUMN  `total_idle_time_sec` INT DEFAULT 0 AFTER `max_idle_duration_sec`,
  ADD COLUMN  `idle_count` INT DEFAULT 0 AFTER `total_idle_time_sec`,
  ADD COLUMN  `workers_created` INT DEFAULT 0 AFTER `idle_count`,
  ADD COLUMN  `workers_terminated` INT DEFAULT 0 AFTER `workers_created`,
  ADD COLUMN  `p50_execution_ms` DECIMAL(10,2) DEFAULT 0 AFTER `avg_execution_ms`;

-- 3. Add missing fields to endpoint_hourly_stats
ALTER TABLE `endpoint_hourly_stats`
  ADD COLUMN  `tasks_retried` INT DEFAULT 0 AFTER `tasks_timeout`,
  ADD COLUMN  `webhook_success` INT DEFAULT 0 AFTER `avg_cold_start_ms`,
  ADD COLUMN  `webhook_failed` INT DEFAULT 0 AFTER `webhook_success`,
  ADD COLUMN  `avg_worker_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `idle_workers`,
  ADD COLUMN  `max_gpu_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `avg_gpu_utilization`,
  ADD COLUMN  `avg_idle_duration_sec` DECIMAL(10,2) DEFAULT 0 AFTER `max_gpu_utilization`,
  ADD COLUMN  `max_idle_duration_sec` INT DEFAULT 0 AFTER `avg_idle_duration_sec`,
  ADD COLUMN  `total_idle_time_sec` BIGINT DEFAULT 0 AFTER `max_idle_duration_sec`,
  ADD COLUMN  `idle_count` INT DEFAULT 0 AFTER `total_idle_time_sec`,
  ADD COLUMN  `workers_created` INT DEFAULT 0 AFTER `idle_count`,
  ADD COLUMN  `workers_terminated` INT DEFAULT 0 AFTER `workers_created`,
  ADD COLUMN  `p50_execution_ms` DECIMAL(10,2) DEFAULT 0 AFTER `avg_execution_ms`;

-- 4. Add missing fields to endpoint_daily_stats
ALTER TABLE `endpoint_daily_stats`
  ADD COLUMN  `tasks_retried` INT DEFAULT 0 AFTER `tasks_timeout`,
  ADD COLUMN  `webhook_success` INT DEFAULT 0 AFTER `avg_cold_start_ms`,
  ADD COLUMN  `webhook_failed` INT DEFAULT 0 AFTER `webhook_success`,
  ADD COLUMN  `avg_worker_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `idle_workers`,
  ADD COLUMN  `max_gpu_utilization` DECIMAL(5,2) DEFAULT 0 AFTER `avg_gpu_utilization`,
  ADD COLUMN  `avg_idle_duration_sec` DECIMAL(10,2) DEFAULT 0 AFTER `max_gpu_utilization`,
  ADD COLUMN  `max_idle_duration_sec` INT DEFAULT 0 AFTER `avg_idle_duration_sec`,
  ADD COLUMN  `total_idle_time_sec` BIGINT DEFAULT 0 AFTER `max_idle_duration_sec`,
  ADD COLUMN  `idle_count` INT DEFAULT 0 AFTER `total_idle_time_sec`,
  ADD COLUMN  `workers_created` INT DEFAULT 0 AFTER `idle_count`,
  ADD COLUMN  `workers_terminated` INT DEFAULT 0 AFTER `workers_created`,
  ADD COLUMN  `p50_execution_ms` DECIMAL(10,2) DEFAULT 0 AFTER `avg_execution_ms`;
