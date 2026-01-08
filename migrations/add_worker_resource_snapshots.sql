-- Worker Resource Snapshots table for monitoring
CREATE TABLE IF NOT EXISTS worker_resource_snapshots (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    worker_id VARCHAR(255) NOT NULL,
    snapshot_at TIMESTAMP NOT NULL,
    gpu_utilization DECIMAL(5,2),
    gpu_memory_used_mb BIGINT,
    gpu_memory_total_mb BIGINT,
    cpu_utilization DECIMAL(5,2),
    memory_used_mb BIGINT,
    memory_total_mb BIGINT,
    current_task_id VARCHAR(255),
    is_idle BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_worker_snapshot (worker_id, snapshot_at),
    INDEX idx_snapshot_time (snapshot_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
