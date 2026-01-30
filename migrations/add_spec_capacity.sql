-- 规格容量状态表
CREATE TABLE IF NOT EXISTS spec_capacity (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    spec_name VARCHAR(100) NOT NULL,
    
    -- 容量状态
    status ENUM('available', 'limited', 'sold_out') NOT NULL DEFAULT 'available',
    reason VARCHAR(255),
    
    -- 统计
    running_count INT NOT NULL DEFAULT 0,
    pending_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    
    -- Spot 信息
    spot_score INT,                          -- Spot Placement Score (1-10)
    spot_price DECIMAL(10,6),                -- 当前 Spot 价格 (USD/hour)
    instance_type VARCHAR(50),               -- 主要实例类型
    
    -- 时间
    last_success_at DATETIME(3),
    last_failure_at DATETIME(3),
    last_spot_check_at DATETIME(3),          -- 上次 Spot 检查时间
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    
    UNIQUE KEY idx_spec_name (spec_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 为现有 spec 初始化容量记录
INSERT INTO spec_capacity (spec_name, status)
SELECT name, 'available' FROM resource_specs WHERE status = 'active'
ON DUPLICATE KEY UPDATE spec_name = spec_name;
