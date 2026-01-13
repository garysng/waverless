-- Migration: Add workers table and monitoring tables, remove GPU usage tables
-- Date: 2026-01-06

-- 1. Create workers table
CREATE TABLE IF NOT EXISTS `workers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `worker_id` varchar(255) NOT NULL COMMENT 'Worker unique ID (usually pod name)',
  `endpoint` varchar(255) NOT NULL COMMENT 'Endpoint name',
  `pod_name` varchar(255) DEFAULT NULL COMMENT 'K8s pod name',
  `status` varchar(50) NOT NULL DEFAULT 'ONLINE' COMMENT 'Worker status: ONLINE, OFFLINE, BUSY, DRAINING',
  `concurrency` int NOT NULL DEFAULT '1' COMMENT 'Maximum concurrency',
  `current_jobs` int NOT NULL DEFAULT '0' COMMENT 'Current number of jobs',
  `jobs_in_progress` text DEFAULT NULL COMMENT 'JSON array of task IDs currently being processed',
  `version` varchar(100) DEFAULT NULL COMMENT 'Worker version',
  `pod_created_at` datetime(3) DEFAULT NULL COMMENT 'Pod creation time',
  `pod_started_at` datetime(3) DEFAULT NULL COMMENT 'Pod started time (container running)',
  `pod_ready_at` datetime(3) DEFAULT NULL COMMENT 'Pod ready time',
  `cold_start_duration_ms` bigint DEFAULT NULL COMMENT 'Cold start duration in milliseconds',
  `last_heartbeat` datetime(3) NOT NULL COMMENT 'Last heartbeat time',
  `last_task_time` datetime(3) DEFAULT NULL COMMENT 'Last task completion time',
  `total_tasks_completed` bigint NOT NULL DEFAULT '0' COMMENT 'Total completed tasks',
  `total_tasks_failed` bigint NOT NULL DEFAULT '0' COMMENT 'Total failed tasks',
  `total_execution_time_ms` bigint NOT NULL DEFAULT '0' COMMENT 'Total execution time in milliseconds',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_worker_id_unique` (`worker_id`),
  KEY `idx_endpoint` (`endpoint`),
  KEY `idx_status` (`status`),
  KEY `idx_last_heartbeat` (`last_heartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Worker records';

-- 2. Add webhook_status to tasks table if not exists
ALTER TABLE `tasks` ADD COLUMN IF NOT EXISTS `webhook_status` varchar(50) DEFAULT NULL COMMENT 'Webhook status: PENDING, SUCCESS, FAILED';

-- 3. Add performance metrics to task_events table if not exists
ALTER TABLE `task_events` ADD COLUMN IF NOT EXISTS `queue_wait_ms` int DEFAULT NULL COMMENT 'Queue wait time in milliseconds';
ALTER TABLE `task_events` ADD COLUMN IF NOT EXISTS `execution_duration_ms` int DEFAULT NULL COMMENT 'Execution duration in milliseconds';
ALTER TABLE `task_events` ADD COLUMN IF NOT EXISTS `total_duration_ms` int DEFAULT NULL COMMENT 'Total duration in milliseconds';

-- 4. Create monitoring aggregation tables
CREATE TABLE IF NOT EXISTS `endpoint_minute_stats` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) NOT NULL,
  `stat_minute` datetime NOT NULL COMMENT 'Minute timestamp',
  `active_workers` int DEFAULT 0,
  `idle_workers` int DEFAULT 0,
  `tasks_submitted` int DEFAULT 0,
  `tasks_completed` int DEFAULT 0,
  `tasks_failed` int DEFAULT 0,
  `tasks_timeout` int DEFAULT 0,
  `avg_queue_wait_ms` decimal(10,2) DEFAULT 0,
  `avg_execution_ms` decimal(10,2) DEFAULT 0,
  `p95_execution_ms` decimal(10,2) DEFAULT 0,
  `avg_gpu_utilization` decimal(5,2) DEFAULT 0,
  `cold_starts` int DEFAULT 0,
  `avg_cold_start_ms` decimal(10,2) DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_endpoint_minute` (`endpoint`, `stat_minute`),
  KEY `idx_stat_minute` (`stat_minute`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Endpoint minute-level statistics';

CREATE TABLE IF NOT EXISTS `endpoint_hourly_stats` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) NOT NULL,
  `stat_hour` datetime NOT NULL COMMENT 'Hour timestamp',
  `active_workers` int DEFAULT 0,
  `idle_workers` int DEFAULT 0,
  `tasks_submitted` int DEFAULT 0,
  `tasks_completed` int DEFAULT 0,
  `tasks_failed` int DEFAULT 0,
  `tasks_timeout` int DEFAULT 0,
  `avg_queue_wait_ms` decimal(10,2) DEFAULT 0,
  `avg_execution_ms` decimal(10,2) DEFAULT 0,
  `p95_execution_ms` decimal(10,2) DEFAULT 0,
  `avg_gpu_utilization` decimal(5,2) DEFAULT 0,
  `cold_starts` int DEFAULT 0,
  `avg_cold_start_ms` decimal(10,2) DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_endpoint_hour` (`endpoint`, `stat_hour`),
  KEY `idx_stat_hour` (`stat_hour`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Endpoint hourly statistics';

CREATE TABLE IF NOT EXISTS `endpoint_daily_stats` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) NOT NULL,
  `stat_date` date NOT NULL COMMENT 'Date',
  `active_workers` int DEFAULT 0,
  `idle_workers` int DEFAULT 0,
  `tasks_submitted` int DEFAULT 0,
  `tasks_completed` int DEFAULT 0,
  `tasks_failed` int DEFAULT 0,
  `tasks_timeout` int DEFAULT 0,
  `avg_queue_wait_ms` decimal(10,2) DEFAULT 0,
  `avg_execution_ms` decimal(10,2) DEFAULT 0,
  `p95_execution_ms` decimal(10,2) DEFAULT 0,
  `avg_gpu_utilization` decimal(5,2) DEFAULT 0,
  `cold_starts` int DEFAULT 0,
  `avg_cold_start_ms` decimal(10,2) DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_endpoint_date` (`endpoint`, `stat_date`),
  KEY `idx_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Endpoint daily statistics';

CREATE TABLE IF NOT EXISTS `worker_resource_snapshots` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `worker_id` varchar(255) NOT NULL,
  `snapshot_at` datetime(3) NOT NULL,
  `gpu_utilization` decimal(5,2) DEFAULT NULL,
  `gpu_memory_used_mb` int DEFAULT NULL,
  `gpu_memory_total_mb` int DEFAULT NULL,
  `gpu_temperature` int DEFAULT NULL,
  `cpu_utilization` decimal(5,2) DEFAULT NULL,
  `memory_used_mb` int DEFAULT NULL,
  `memory_total_mb` int DEFAULT NULL,
  `current_task_id` varchar(255) DEFAULT NULL,
  `is_idle` tinyint(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_worker_snapshot` (`worker_id`, `snapshot_at`),
  KEY `idx_snapshot_at` (`snapshot_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Worker resource usage snapshots';

-- 5. Incremental ALTER statements for existing databases
-- Add jobs_in_progress to workers if not exists
ALTER TABLE `workers` ADD COLUMN IF NOT EXISTS `jobs_in_progress` text DEFAULT NULL COMMENT 'JSON array of task IDs currently being processed' AFTER `current_jobs`;

-- Add cold_starts and avg_cold_start_ms to endpoint_minute_stats if not exists
ALTER TABLE `endpoint_minute_stats` ADD COLUMN IF NOT EXISTS `cold_starts` int DEFAULT 0 AFTER `avg_gpu_utilization`;
ALTER TABLE `endpoint_minute_stats` ADD COLUMN IF NOT EXISTS `avg_cold_start_ms` decimal(10,2) DEFAULT 0 AFTER `cold_starts`;

-- 6. Drop old GPU usage tables (optional - uncomment if you want to remove old data)
-- DROP TABLE IF EXISTS `gpu_usage_records`;
-- DROP TABLE IF EXISTS `gpu_usage_statistics_minute`;
-- DROP TABLE IF EXISTS `gpu_usage_statistics_hourly`;
-- DROP TABLE IF EXISTS `gpu_usage_statistics_daily`;
