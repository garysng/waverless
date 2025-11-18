

CREATE TABLE `autoscaler_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name',
  `display_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Display name',
  `spec_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Spec name for resource calculation',
  `min_replicas` int NOT NULL DEFAULT '0' COMMENT 'Minimum replica count',
  `max_replicas` int NOT NULL DEFAULT '10' COMMENT 'Maximum replica count',
  `replicas` int NOT NULL DEFAULT '1' COMMENT 'Target replica count',
  `scale_up_threshold` int NOT NULL DEFAULT '1' COMMENT 'Queue length threshold for scale up',
  `scale_down_idle_time` int NOT NULL DEFAULT '300' COMMENT 'Idle time in seconds before scale down',
  `scale_up_cooldown` int NOT NULL DEFAULT '30' COMMENT 'Scale up cooldown in seconds',
  `scale_down_cooldown` int NOT NULL DEFAULT '60' COMMENT 'Scale down cooldown in seconds',
  `priority` int NOT NULL DEFAULT '50' COMMENT 'Base priority (0-100)',
  `enable_dynamic_prio` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Enable dynamic priority adjustment',
  `high_load_threshold` int NOT NULL DEFAULT '10' COMMENT 'High load threshold for priority boost',
  `priority_boost` int NOT NULL DEFAULT '20' COMMENT 'Priority boost amount for high load',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Whether autoscaling is enabled for this endpoint',
  `autoscaler_enabled` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Autoscaler override: NULL/"" = follow global, "disabled" = force off, "enabled" = force on',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `last_task_time` datetime(3) DEFAULT NULL COMMENT 'Last task completion time (for idle time calculation)',
  `last_scale_time` datetime(3) DEFAULT NULL COMMENT 'Last scaling operation time (for cooldown)',
  `first_pending_time` datetime(3) DEFAULT NULL COMMENT 'First pending task time (for starvation prevention)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_endpoint_unique` (`endpoint`),
  KEY `idx_enabled` (`enabled`),
  KEY `idx_last_task_time` (`last_task_time`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Autoscaler configuration per endpoint';

CREATE TABLE `endpoints` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name (unique identifier)',
  `spec_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Resource spec name',
  `image` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Docker image',
  `replicas` int NOT NULL DEFAULT '1' COMMENT 'Target replica count',
  `task_timeout` int NOT NULL DEFAULT '0' COMMENT 'Task execution timeout in seconds (0 = use global default)',
  `env` json DEFAULT NULL COMMENT 'Environment variables as JSON object',
  `labels` json DEFAULT NULL COMMENT 'Labels as JSON object',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'active' COMMENT 'Endpoint status: active, inactive, deleted',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_endpoint_unique` (`endpoint`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Endpoint metadata and deployment configuration';

ALTER TABLE `endpoints` ADD COLUMN `enable_ptrace` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Enable SYS_PTRACE capability for debugging (only for fixed resource pools)';

CREATE TABLE `gpu_usage_records` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Task ID',
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name',
  `worker_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Worker/Pod ID',
  `spec_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Spec name (e.g., h200-single, NVIDIA-A100-80GB)',
  `gpu_count` int NOT NULL DEFAULT '1' COMMENT 'Number of GPUs used',
  `gpu_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'GPU type (e.g., NVIDIA-H200, A100)',
  `gpu_memory_gb` int DEFAULT NULL COMMENT 'GPU memory per card (GB)',
  `started_at` datetime(3) NOT NULL COMMENT 'Task start time',
  `completed_at` datetime(3) NOT NULL COMMENT 'Task completion time',
  `duration_seconds` int NOT NULL COMMENT 'Task duration in seconds',
  `gpu_hours` decimal(10,4) NOT NULL COMMENT 'GPU card-hours (gpu_count × duration_hours)',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Task status (COMPLETED, FAILED)',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_task_id_unique` (`task_id`),
  KEY `idx_endpoint` (`endpoint`),
  KEY `idx_spec_name` (`spec_name`),
  KEY `idx_started_at` (`started_at`),
  KEY `idx_completed_at` (`completed_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7115 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='GPU usage records at task level';

CREATE TABLE `gpu_usage_statistics_daily` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `time_bucket` date NOT NULL COMMENT 'Time bucket (day granularity, e.g., 2025-11-12)',
  `scope_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Scope type: global, endpoint, or spec',
  `scope_value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Endpoint/Spec name (NULL for global)',
  `total_tasks` int DEFAULT '0' COMMENT 'Total tasks completed in this day',
  `completed_tasks` int DEFAULT '0' COMMENT 'Successfully completed tasks',
  `failed_tasks` int DEFAULT '0' COMMENT 'Failed tasks',
  `total_gpu_hours` decimal(10,4) DEFAULT '0.0000' COMMENT 'Total GPU card-hours used',
  `avg_gpu_count` decimal(10,2) DEFAULT '0.00' COMMENT 'Average GPU count per task',
  `max_gpu_count` int DEFAULT '0' COMMENT 'Max GPU count in any single task',
  `available_gpu_hours` decimal(10,4) DEFAULT NULL COMMENT 'Available GPU hours (capacity)',
  `utilization_rate` decimal(5,2) DEFAULT NULL COMMENT 'GPU utilization rate (%)',
  `peak_hour` datetime DEFAULT NULL COMMENT 'Hour with highest GPU usage',
  `peak_gpu_hours` decimal(10,4) DEFAULT NULL COMMENT 'GPU hours in peak hour',
  `period_start` datetime(3) NOT NULL COMMENT 'Period start time',
  `period_end` datetime(3) NOT NULL COMMENT 'Period end time',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `scope_value_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci GENERATED ALWAYS AS (coalesce(`scope_value`,_utf8mb4'__GLOBAL__')) STORED,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_time_scope` (`time_bucket`,`scope_type`,`scope_value_key`),
  KEY `idx_time_bucket` (`time_bucket`),
  KEY `idx_scope` (`scope_type`,`scope_value`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB AUTO_INCREMENT=103 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Daily aggregated GPU usage statistics';

CREATE TABLE `gpu_usage_statistics_hourly` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `time_bucket` datetime NOT NULL COMMENT 'Time bucket (hour granularity, e.g., 2025-11-12 10:00:00)',
  `scope_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Scope type: global, endpoint, or spec',
  `scope_value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Endpoint/Spec name (NULL for global)',
  `total_tasks` int DEFAULT '0' COMMENT 'Total tasks completed in this hour',
  `completed_tasks` int DEFAULT '0' COMMENT 'Successfully completed tasks',
  `failed_tasks` int DEFAULT '0' COMMENT 'Failed tasks',
  `total_gpu_hours` decimal(10,4) DEFAULT '0.0000' COMMENT 'Total GPU card-hours used',
  `avg_gpu_count` decimal(10,2) DEFAULT '0.00' COMMENT 'Average GPU count per task',
  `max_gpu_count` int DEFAULT '0' COMMENT 'Max GPU count in any single task',
  `peak_minute` datetime DEFAULT NULL COMMENT 'Minute with highest GPU usage',
  `peak_gpu_hours` decimal(10,4) DEFAULT NULL COMMENT 'GPU hours in peak minute',
  `period_start` datetime(3) NOT NULL COMMENT 'Period start time',
  `period_end` datetime(3) NOT NULL COMMENT 'Period end time',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `scope_value_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci GENERATED ALWAYS AS (coalesce(`scope_value`,_utf8mb4'__GLOBAL__')) STORED,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_time_scope` (`time_bucket`,`scope_type`,`scope_value_key`),
  KEY `idx_time_bucket` (`time_bucket`),
  KEY `idx_scope` (`scope_type`,`scope_value`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB AUTO_INCREMENT=377 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Hourly aggregated GPU usage statistics';

CREATE TABLE `gpu_usage_statistics_minute` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `time_bucket` datetime NOT NULL COMMENT 'Time bucket (minute granularity, e.g., 2025-11-12 10:30:00)',
  `scope_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Scope type: global, endpoint, or spec',
  `scope_value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Endpoint/Spec name (NULL for global)',
  `total_tasks` int DEFAULT '0' COMMENT 'Total tasks completed in this minute',
  `completed_tasks` int DEFAULT '0' COMMENT 'Successfully completed tasks',
  `failed_tasks` int DEFAULT '0' COMMENT 'Failed tasks',
  `total_gpu_seconds` decimal(12,2) DEFAULT '0.00' COMMENT 'Total GPU card-seconds used',
  `total_gpu_hours` decimal(10,4) DEFAULT '0.0000' COMMENT 'Total GPU card-hours used',
  `avg_gpu_count` decimal(10,2) DEFAULT '0.00' COMMENT 'Average GPU count per task',
  `max_gpu_count` int DEFAULT '0' COMMENT 'Max GPU count in any single task',
  `period_start` datetime(3) NOT NULL COMMENT 'Period start time',
  `period_end` datetime(3) NOT NULL COMMENT 'Period end time',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `scope_value_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci GENERATED ALWAYS AS (coalesce(`scope_value`,_utf8mb4'__GLOBAL__')) STORED,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_time_scope` (`time_bucket`,`scope_type`,`scope_value_key`),
  KEY `idx_time_bucket` (`time_bucket`),
  KEY `idx_scope` (`scope_type`,`scope_value`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB AUTO_INCREMENT=6496 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Minute-level aggregated GPU usage statistics';

CREATE TABLE `scaling_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Event unique ID',
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name',
  `timestamp` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `action` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Action: scale_up, scale_down, blocked, preempted',
  `from_replicas` int NOT NULL COMMENT 'Original replica count',
  `to_replicas` int NOT NULL COMMENT 'Target replica count',
  `reason` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Reason for this scaling action',
  `queue_length` bigint NOT NULL DEFAULT '0' COMMENT 'Pending task queue length',
  `priority` int NOT NULL DEFAULT '50' COMMENT 'Effective priority at the time',
  `preempted_from` json DEFAULT NULL COMMENT 'List of endpoints this action preempted from',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_event_id_unique` (`event_id`),
  KEY `idx_endpoint_timestamp` (`endpoint`,`timestamp`),
  KEY `idx_action` (`action`),
  KEY `idx_timestamp` (`timestamp`)
) ENGINE=InnoDB AUTO_INCREMENT=7815 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Autoscaling event history';

CREATE TABLE `task_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Event unique ID',
  `task_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Task ID (foreign key to tasks.task_id)',
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name',
  `event_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Event type: TASK_CREATED, TASK_ASSIGNED, TASK_COMPLETED, etc.',
  `event_time` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Event timestamp',
  `worker_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Worker ID',
  `worker_pod_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Worker Pod name in Kubernetes',
  `from_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Original status',
  `to_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'New status after this event',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT 'Error message if event is failure-related',
  `error_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Error type classification',
  `retry_count` int NOT NULL DEFAULT '0' COMMENT 'Retry count at the time of this event',
  `metadata` json DEFAULT NULL COMMENT 'Additional event metadata',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_event_id_unique` (`event_id`),
  KEY `idx_task_id_event_time` (`task_id`,`event_time`),
  KEY `idx_endpoint_event_time` (`endpoint`,`event_time`),
  KEY `idx_worker_id` (`worker_id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_event_time` (`event_time`)
) ENGINE=InnoDB AUTO_INCREMENT=105022 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Task event log for detailed tracking and auditing';

CREATE TABLE `task_statistics` (
  `id` int NOT NULL AUTO_INCREMENT,
  `scope_type` varchar(50) NOT NULL COMMENT 'Statistics scope: global or endpoint',
  `scope_value` varchar(255) DEFAULT NULL COMMENT 'Endpoint name (NULL for global scope)',
  `pending_count` int DEFAULT '0' COMMENT 'Number of PENDING tasks',
  `in_progress_count` int DEFAULT '0' COMMENT 'Number of IN_PROGRESS tasks',
  `completed_count` int DEFAULT '0' COMMENT 'Number of COMPLETED tasks',
  `failed_count` int DEFAULT '0' COMMENT 'Number of FAILED tasks',
  `cancelled_count` int DEFAULT '0' COMMENT 'Number of CANCELLED tasks',
  `total_count` int DEFAULT '0' COMMENT 'Total number of tasks',
  `updated_at` datetime(3) NOT NULL COMMENT 'Last update timestamp',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scope` (`scope_type`,`scope_value`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB AUTO_INCREMENT=77671 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Task statistics for dashboard';

CREATE TABLE `tasks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Task unique ID (UUID)',
  `endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Endpoint name',
  `input` json NOT NULL COMMENT 'Task input parameters as JSON',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Task status: PENDING, IN_PROGRESS, COMPLETED, FAILED, CANCELLED',
  `output` json DEFAULT NULL COMMENT 'Task output as JSON',
  `error` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT 'Error message if task failed',
  `worker_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Worker ID processing this task',
  `webhook_url` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Webhook URL for completion notification',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `started_at` datetime(3) DEFAULT NULL COMMENT 'Time when task started processing',
  `completed_at` datetime(3) DEFAULT NULL COMMENT 'Time when task completed (success or failure)',
  `extend` json DEFAULT NULL COMMENT 'Execution history summary and extended info',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_task_id_unique` (`task_id`),
  KEY `idx_endpoint_status` (`endpoint`,`status`),
  KEY `idx_status` (`status`),
  KEY `idx_worker_id` (`worker_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_completed_at` (`completed_at`)
) ENGINE=InnoDB AUTO_INCREMENT=26304 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Task records with all statuses';

ALTER TABLE `tasks`
  ADD INDEX `idx_endpoint_id` (`endpoint`, `id`),
  ADD INDEX `idx_endpoint_status_id` (`endpoint`, `status`, `id`),
  ADD INDEX `idx_status_id` (`status`, `id`);
  
CREATE TABLE `resource_specs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Spec name (unique identifier)',
  `display_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Display name',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Category: cpu, gpu',
  `cpu` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'CPU cores (e.g., "2", "4")',
  `memory` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Memory (e.g., "4Gi", "8Gi")',
  `gpu` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'GPU count (e.g., "1", "2")',
  `gpu_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'GPU type (e.g., "NVIDIA-H200", "NVIDIA-A100")',
  `ephemeral_storage` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Ephemeral storage (e.g., "30", "300")',
  `shm_size` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Shared memory size (e.g., "1Gi", "512Mi")',
  `resource_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'serverless' COMMENT 'Resource type: fixed, serverless',
  `platforms` json DEFAULT NULL COMMENT 'Platform-specific configurations as JSON',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'active' COMMENT 'Spec status: active, inactive, deprecated',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_spec_name_unique` (`name`),
  KEY `idx_category` (`category`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Resource specifications for deployments';

