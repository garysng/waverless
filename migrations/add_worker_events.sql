-- Migration: Add worker_events table for worker lifecycle tracking
-- Date: 2026-01-08

-- Worker 事件表（记录 Worker 生命周期事件）
CREATE TABLE IF NOT EXISTS `worker_events` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `event_id` VARCHAR(255) NOT NULL UNIQUE COMMENT 'Unique event ID',
  `worker_id` VARCHAR(255) NOT NULL COMMENT 'Worker ID',
  `endpoint` VARCHAR(255) NOT NULL COMMENT 'Endpoint name',
  `event_type` VARCHAR(50) NOT NULL COMMENT 'Event type: WORKER_STARTED, WORKER_REGISTERED, WORKER_TASK_PULLED, WORKER_TASK_COMPLETED, WORKER_OFFLINE, WORKER_IDLE',
  `event_time` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Event timestamp',
  
  -- 冷启动相关
  `cold_start_duration_ms` BIGINT DEFAULT NULL COMMENT 'Cold start duration (for WORKER_REGISTERED event)',
  
  -- 空闲相关
  `idle_duration_ms` BIGINT DEFAULT NULL COMMENT 'Idle duration before this event (for WORKER_TASK_PULLED event)',
  
  -- 任务相关
  `task_id` VARCHAR(255) DEFAULT NULL COMMENT 'Related task ID (for task events)',
  
  -- 扩展信息
  `metadata` JSON DEFAULT NULL COMMENT 'Additional metadata',
  
  INDEX `idx_worker_event_time` (`worker_id`, `event_time`),
  INDEX `idx_endpoint_event_time` (`endpoint`, `event_time`),
  INDEX `idx_event_type_time` (`event_type`, `event_time`),
  INDEX `idx_event_time` (`event_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Worker lifecycle events';

-- 删除 webhook 相关字段（暂不需要）
ALTER TABLE `endpoint_minute_stats` DROP COLUMN IF EXISTS `webhook_success`;
ALTER TABLE `endpoint_minute_stats` DROP COLUMN IF EXISTS `webhook_failed`;
ALTER TABLE `endpoint_hourly_stats` DROP COLUMN IF EXISTS `webhook_success`;
ALTER TABLE `endpoint_hourly_stats` DROP COLUMN IF EXISTS `webhook_failed`;
ALTER TABLE `endpoint_daily_stats` DROP COLUMN IF EXISTS `webhook_success`;
ALTER TABLE `endpoint_daily_stats` DROP COLUMN IF EXISTS `webhook_failed`;

-- 可以删除 worker_resource_snapshots 表（用 worker_events 替代）
-- DROP TABLE IF EXISTS `worker_resource_snapshots`;
