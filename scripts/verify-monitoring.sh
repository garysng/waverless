#!/bin/bash

ENDPOINT="wan-2-2-i2v-14b-no-lora"
API_BASE="http://localhost:8080"

echo "=== 监控指标验证脚本 ==="
echo ""

# 1. 查看实时指标
echo "1. 实时指标 (Worker Utilization 来源)"
curl -s "$API_BASE/v1/$ENDPOINT/metrics/realtime" | jq '.workers'
echo ""

# 2. 查看 Workers 状态
echo "2. Workers 状态"
mysql -h sg-cdb-qmk985xh.sql.tencentcdb.com -P 63938 -u wavespeed -p'wavespeed&123!' waverless -e "
SELECT worker_id, status, current_jobs, last_heartbeat 
FROM workers 
WHERE endpoint = '$ENDPOINT' 
ORDER BY last_heartbeat DESC;" 2>&1 | grep -v "Warning"
echo ""

# 3. 查看资源快照 (Idle Time/Frequency 来源)
echo "3. 资源快照数据 (最近10条)"
mysql -h sg-cdb-qmk985xh.sql.tencentcdb.com -P 63938 -u wavespeed -p'wavespeed&123!' waverless -e "
SELECT worker_id, snapshot_at, is_idle, current_task_id 
FROM worker_resource_snapshots 
WHERE endpoint = '$ENDPOINT' 
ORDER BY snapshot_at DESC 
LIMIT 10;" 2>&1 | grep -v "Warning"
echo ""

# 4. 查看最新的分钟级统计
echo "4. 最新分钟级统计"
mysql -h sg-cdb-qmk985xh.sql.tencentcdb.com -P 63938 -u wavespeed -p'wavespeed&123!' waverless -e "
SELECT stat_minute, active_workers, idle_workers, avg_worker_utilization, 
       avg_idle_duration_sec, idle_count, workers_created, workers_terminated
FROM endpoint_minute_stats 
WHERE endpoint = '$ENDPOINT' 
ORDER BY stat_minute DESC 
LIMIT 5;" 2>&1 | grep -v "Warning"
echo ""

# 5. 提交测试任务
echo "5. 提交测试任务..."
TASK_ID=$(curl -s -X POST "$API_BASE/v1/$ENDPOINT/run" \
  -H "Content-Type: application/json" \
  -d '{"input": {"prompt": "test monitoring"}}' | jq -r '.id')
echo "Task ID: $TASK_ID"
echo ""

# 6. 等待10秒后再次查看
echo "6. 等待10秒后查看 Worker 状态..."
sleep 10
mysql -h sg-cdb-qmk985xh.sql.tencentcdb.com -P 63938 -u wavespeed -p'wavespeed&123!' waverless -e "
SELECT worker_id, status, current_jobs, last_heartbeat 
FROM workers 
WHERE endpoint = '$ENDPOINT' 
ORDER BY last_heartbeat DESC;" 2>&1 | grep -v "Warning"
echo ""

echo "=== 验证完成 ==="
echo ""
echo "预期结果:"
echo "- Worker Utilization: 如果有任务在执行，current_jobs > 0，利用率应该 > 0"
echo "- Worker Idle Time: 需要等待采集器运行后才有数据"
echo "- Worker Idle Frequency: 需要等待采集器运行后才有数据"
echo "- Worker Lifecycle: 如果有扩缩容，会看到 workers_created/terminated > 0"
