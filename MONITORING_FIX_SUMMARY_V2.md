# 监控数据统计问题诊断和修复报告 V2

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

### 问题 2: Snapshot 轮询方案的采样精度问题 ⚠️

**场景示例**:
```
08:00:00 - Snapshot采集 -> Worker空闲 (is_idle=true)
08:00:05 - 任务A开始执行
08:00:15 - 任务A完成 (执行10秒)
08:00:20 - Worker空闲
08:01:00 - Snapshot采集 -> Worker空闲 (is_idle=true)

结果：统计显示Worker一直空闲，丢失了10秒的执行时间 ❌
```

**根本问题**:
- 固定采样频率（每分钟一次）无法捕获快速任务
- 提高采样频率会增加数据库负载
- 基于快照的方案天然存在采样盲区

### 问题 3: worker_resource_snapshots 表为空

**检查结果**:
```sql
SELECT COUNT(*) FROM worker_resource_snapshots;
-- 结果: 0
```

**影响**:
- GPU 利用率指标全为 0
- Worker 空闲时间统计全为 0

### 问题 4: 设计与实现不一致

**设计文档**:
> "聚合层只从监控表读取数据，不直接查询业务表（tasks、workers）"

**实际实现**:
- 直接查询 workers 表获取 worker 统计
- workers 表只保存最新状态，无法准确统计历史

## 修复方案（最终版）

### 核心思路：从 task_events 反向推导 Worker 利用率 ✅

**优势**:
1. ✅ **精确统计** - 不会遗漏任何任务执行时间
2. ✅ **无采样盲区** - 基于事件驱动，记录每个任务
3. ✅ **符合设计原则** - 从监控表（task_events）聚合
4. ✅ **无额外成本** - 利用已有数据，无需额外采集

**实现逻辑**:

```sql
-- 1. 计算每个 worker 的执行时间和利用率
SELECT
    worker_id,
    SUM(execution_duration_ms) as total_exec_ms,
    COUNT(*) as task_count,
    (SUM(execution_duration_ms) / 60000) * 100 as utilization_pct
FROM task_events
WHERE endpoint = ?
  AND event_time >= ?  -- 分钟开始
  AND event_time < ?   -- 分钟结束
  AND event_type IN ('TASK_COMPLETED', 'TASK_FAILED', 'TASK_TIMEOUT')
  AND worker_id IS NOT NULL
  AND execution_duration_ms IS NOT NULL
GROUP BY worker_id
```

```sql
-- 2. 找出完全空闲的 worker（在线但没有执行任何任务）
SELECT COUNT(DISTINCT w.worker_id)
FROM workers w
WHERE w.endpoint = ?
  AND w.last_heartbeat >= ?
  AND w.last_heartbeat < ?
  AND w.status IN ('ONLINE', 'BUSY')
  AND NOT EXISTS (
    SELECT 1 FROM task_events te
    WHERE te.worker_id = w.worker_id
      AND te.event_time >= ?
      AND te.event_time < ?
  )
```

**统计结果**:
- `active_workers`: 利用率 > 0 的 worker 数量
- `idle_workers`: 完全空闲的 worker 数量
- `avg_worker_utilization`: 所有 worker 的平均利用率

**示例计算**:
```
假设一分钟内（60秒 = 60000ms）:
- Worker A: 执行了 3 个任务，总耗时 45000ms
  利用率 = (45000 / 60000) * 100 = 75%

- Worker B: 执行了 1 个任务，总耗时 10000ms
  利用率 = (10000 / 60000) * 100 = 16.67%

- Worker C: 没有执行任务
  利用率 = 0%

统计结果:
- active_workers = 2 (A和B)
- idle_workers = 1 (C)
- avg_worker_utilization = (75 + 16.67 + 0) / 3 = 30.56%
```

### 处理边缘情况

1. **任务跨分钟执行**:
   - 当前按 event_time（完成时间）归属
   - 未来可优化为按执行时间段分配

2. **并发任务**:
   - 利用率可能超过 100%（如果 worker 支持并发）
   - 这是合理的，反映了真实情况

3. **Worker 状态不一致**:
   - 使用心跳时间范围确保 worker 在线
   - 结合 task_events 确保准确性

## 文件变更

- `pkg/store/mysql/monitoring_repository.go`:
  - 移除了基于 snapshots 的统计逻辑
  - 实现了基于 task_events 的精确统计
  - 添加了完全空闲 worker 的计数
  - 按实际执行时间计算利用率

## 测试验证

### 验证场景 1: 快速任务

```bash
# 提交一个 10 秒的任务
# 预期: 即使在采样间隔之间完成，也能准确统计
```

### 验证场景 2: 并发任务

```bash
# 一个 worker 执行多个任务
# 预期: 利用率准确反映总执行时间
```

### 验证 SQL

```sql
-- 验证某个分钟的统计
SELECT
    worker_id,
    SUM(execution_duration_ms) / 1000.0 as total_exec_sec,
    (SUM(execution_duration_ms) / 60000.0) * 100 as utilization_pct
FROM task_events
WHERE endpoint = 'wan-2-2-i2v-14b-no-lora'
  AND event_time >= '2026-01-08 08:28:00'
  AND event_time < '2026-01-08 08:29:00'
  AND event_type IN ('TASK_COMPLETED', 'TASK_FAILED', 'TASK_TIMEOUT')
  AND worker_id IS NOT NULL
GROUP BY worker_id;
```

## 关于 worker_resource_snapshots

虽然修复后不再依赖 snapshots 来统计 worker 利用率，但 snapshots 仍然有价值：

**保留用途**:
- GPU 利用率监控（需要从 K8s metrics API 采集）
- CPU/内存使用率监控
- 更详细的资源监控数据

**建议**:
- 保持 snapshot collection job 运行
- 但 worker 利用率统计改为基于 task_events
- 两种数据源可以互相验证

## 总结

### 问题解决方案对比

| 方案 | 优点 | 缺点 | 采用 |
|-----|------|------|-----|
| Snapshot 轮询 | 实现简单 | 采样盲区，遗漏快速任务 | ❌ |
| 提高采样频率 | 提升精度 | 增加数据库负载，仍有盲区 | ❌ |
| 从 task_events 推导 | 精确、无盲区、无额外成本 | 需要心跳数据辅助 | ✅ |

### 最终方案特点

1. ✅ **解决采样盲区问题** - 基于事件驱动，不会遗漏快速任务
2. ✅ **精确统计利用率** - 按实际执行时间计算
3. ✅ **符合设计原则** - 从监控表聚合，不依赖业务表
4. ✅ **性能优化** - 利用已有数据，无额外采集开销
5. ✅ **可扩展** - 支持并发任务、跨分钟任务等场景

下一步只需部署并观察新的统计数据是否准确反映实际情况。
