# 监控数据统计问题诊断和修复报告

## 问题描述

在 `endpoint_minute_stats` 表中，id=163 的记录显示：
- `tasks_submitted: 1`, `tasks_completed: 1` - 任务数据正常
- `active_workers: 1`, `idle_workers: 1` - worker 数量有记录
- `avg_worker_utilization: 0.00` - 但利用率为 0
- 所有 idle/GPU 相关指标都是 0

## 根本原因分析

### 问题 1: Worker 统计查询逻辑错误

**位置**: `pkg/store/mysql/monitoring_repository.go:215` (原来)

**原因**:
```sql
-- 错误的查询
SELECT ... FROM workers 
WHERE endpoint = ? AND last_heartbeat >= ?
```

只使用了 `last_heartbeat >= from`，缺少上限检查：
1. 会匹配所有在 `from` 时间点之后有过心跳的 worker
2. 查询的是 workers 表的**当前快照状态**，而不是历史状态
3. Worker 状态可能已改变（如 ONLINE → OFFLINE）

**验证**:
```bash
# 查询 id=163 的时间段 (2026-01-08 08:28:00-08:29:00)
# 错误查询返回: 4 个 worker，但都是 OFFLINE 状态
# 正确查询返回: 0 个 worker (因为没有在这一分钟内发送心跳)
```

### 问题 2: worker_resource_snapshots 表为空

**检查结果**:
```sql
SELECT COUNT(*) FROM worker_resource_snapshots;
-- 结果: 0
```

**原因**:
- Snapshot collection job 已经注册（`cmd/jobs.go:63`）
- Collector 已经初始化（`cmd/initializers.go:166`）
- 但表中没有数据，说明可能：
  1. Job 还未开始运行
  2. 有运行时错误阻止数据保存

**影响**:
- GPU 利用率指标全为 0
- Worker 空闲时间统计全为 0
- Worker 利用率统计不准确

### 问题 3: 设计与实现不一致

**设计文档** (`docs/WAVERLESS_MONITORING_DESIGN.md:787`):
> "聚合层只从监控表读取数据，不直接查询业务表（tasks、workers）"

**实际实现**:
- 直接查询 workers 表获取 worker 统计
- 违反了设计原则

**根本问题**:
- workers 表只保存最新状态，不保留历史
- 无法准确统计历史时间段的 worker 状态
- 应该从 `worker_resource_snapshots` 表聚合

## 修复方案

### 修复 1: 添加时间上限检查

```go
// 修复前
FROM workers WHERE endpoint = ? AND last_heartbeat >= ?

// 修复后  
FROM workers WHERE endpoint = ? AND last_heartbeat >= ? AND last_heartbeat < ?
```

**效果**: 只查询在指定时间段内有心跳的 worker

### 修复 2: 优先使用 Snapshots，提供降级方案

**新逻辑**:
1. **优先使用 worker_resource_snapshots** (准确的历史状态)
   ```sql
   SELECT
       COUNT(DISTINCT CASE WHEN NOT is_idle THEN worker_id END) as active_workers,
       COUNT(DISTINCT CASE WHEN is_idle THEN worker_id END) as idle_workers,
       COUNT(DISTINCT worker_id) as total_workers
   FROM worker_resource_snapshots
   WHERE endpoint = ? AND snapshot_at >= ? AND snapshot_at < ?
   ```

2. **降级使用 workers 表** (当 snapshots 没数据时)
   ```sql
   SELECT ... FROM workers 
   WHERE endpoint = ? AND last_heartbeat >= ? AND last_heartbeat < ?
   ```

**好处**:
- 符合设计原则
- 提供准确的历史数据
- 有降级方案保证可用性

## 后续建议

### 1. 确保 Snapshot Collection 正常运行

检查:
```bash
# 查看 job 日志
kubectl logs -n wavespeed-test <pod-name> | grep "snapshot-collection"

# 查看是否有错误
kubectl logs -n wavespeed-test <pod-name> | grep -i "error.*snapshot"
```

### 2. 手动触发一次聚合验证

```bash
# 提交一个测试任务
# 等待 1-2 分钟
# 检查 worker_resource_snapshots 是否有新数据
```

### 3. 监控数据完整性

添加监控告警:
- `worker_resource_snapshots` 表数据采集率
- 监控聚合任务的成功率
- Worker 统计数据的合理性检查

## 文件变更

- `pkg/store/mysql/monitoring_repository.go`:
  - 修复了 worker 统计查询逻辑
  - 添加了从 snapshots 聚合的优先逻辑
  - 保留了 workers 表的降级方案

## 测试验证

### 验证步骤

1. 等待 snapshot collection job 运行（每分钟一次）
2. 检查 `worker_resource_snapshots` 是否有新数据
3. 等待 minute aggregation job 运行
4. 检查新的 `endpoint_minute_stats` 记录中 worker 数据是否正确

### 预期结果

- 如果有 worker 在线且执行任务: `avg_worker_utilization > 0`
- 如果有 worker 空闲: `idle_workers > 0` 且 `total_idle_time_sec > 0`
- 统计数据与实际 worker 状态一致

## 总结

根本问题是 worker 统计逻辑不正确，混用了业务表（workers）和监控表（worker_resource_snapshots）。修复后的实现：

1. ✅ 优先使用 snapshots（设计要求）
2. ✅ 添加降级方案（可用性保证）
3. ✅ 修复时间范围查询（准确性保证）

下一步需要确认 snapshot collection job 正常运行，确保有持续的快照数据采集。
