# Waverless 代码优化计划

## 一、常量定义优化

### 1.1 Worker 状态常量 (高优先级)
**问题**: Worker 状态字符串散落在多处，如 `"ONLINE"`, `"OFFLINE"`, `"BUSY"`, `"STARTING"`, `"DRAINING"`

**现状**:
- `internal/model/worker.go` 定义了 `WorkerStatus` 类型和常量
- 但 `pkg/store/mysql/worker_repository.go` 直接使用字符串 `"ONLINE"`, `"BUSY"`
- `pkg/autoscaler/metrics_collector.go` 使用 `"DRAINING"`

**优化方案**:
```go
// pkg/constants/worker.go
package constants

type WorkerStatus string

const (
    WorkerStatusStarting WorkerStatus = "STARTING"
    WorkerStatusOnline   WorkerStatus = "ONLINE"
    WorkerStatusBusy     WorkerStatus = "BUSY"
    WorkerStatusDraining WorkerStatus = "DRAINING"
    WorkerStatusOffline  WorkerStatus = "OFFLINE"
)
```

### 1.2 Task 状态常量 (高优先级)
**问题**: Task 状态字符串散落在多处，如 `"PENDING"`, `"IN_PROGRESS"`, `"COMPLETED"`, `"FAILED"`

**现状**:
- `internal/model/task.go` 定义了 `TaskStatus` 类型
- 但 `pkg/store/mysql/task_repository.go` 直接使用字符串
- `internal/service/task_service.go` 混用常量和字符串

**优化方案**: 统一使用 `model.TaskStatus*` 常量

### 1.3 K8s Label 常量 (中优先级)
**问题**: K8s label key 硬编码在多处

**现状**:
```go
// 散落在 manager.go, pods.go, jobs.go
pod.Labels["app"]
labels.Set{"app": endpoint}
```

**优化方案**:
```go
// pkg/constants/k8s.go
package constants

const (
    LabelApp       = "app"
    LabelManagedBy = "managed-by"
    LabelComponent = "component"
    
    ManagedByWaverless = "waverless"
)
```

### 1.4 Pod 状态常量 (中优先级)
**问题**: Pod 状态字符串如 `"Running"`, `"Pending"`, `"Terminating"` 硬编码

**优化方案**:
```go
// pkg/constants/pod.go
const (
    PodPhaseRunning    = "Running"
    PodPhasePending    = "Pending"
    PodPhaseSucceeded  = "Succeeded"
    PodPhaseFailed     = "Failed"
    
    PodStatusStarting    = "Starting"
    PodStatusTerminating = "Terminating"
)
```

---

## 二、目录结构优化

### 2.1 worker_lister.go 位置错误 (高优先级)
**问题**: `pkg/autoscaler/worker_lister.go` 不应该在 autoscaler 目录

**现状**: 这是一个通用的 Worker 查询接口，被 autoscaler 使用

**优化方案**: 
- 移动到 `pkg/interfaces/worker.go` 或 `internal/service/worker_lister.go`
- 或者在 `pkg/store/mysql/worker_repository.go` 中实现接口

### 2.2 interfaces 目录整理 (中优先级)
**问题**: `pkg/interfaces/` 目录文件职责不清晰

**现状**:
```
pkg/interfaces/
├── autoscaler.go    # AutoScaler 接口
├── deployment.go    # 部署相关接口和类型（太大）
├── metadata.go      # Endpoint 元数据接口
└── repository.go    # 空文件
```

**优化方案**:
```
pkg/interfaces/
├── deployment.go    # DeploymentProvider 接口
├── endpoint.go      # Endpoint 相关类型
├── worker.go        # Worker 相关接口
├── pod.go           # PodInfo 等类型
└── autoscaler.go    # AutoScaler 接口
```

### 2.3 model 目录整理 (低优先级)
**问题**: 存在两套 model
- `internal/model/` - 业务模型
- `pkg/store/mysql/model/` - 数据库模型

**优化方案**: 保持现状，但确保转换逻辑清晰

---

## 三、重复代码消除

### 3.1 Pod Label 提取逻辑 (高优先级)
**问题**: 多处重复提取 Pod 的 endpoint label

**现状**:
```go
// manager.go - handlePodAdd
endpoint := ""
if pod.Labels != nil {
    endpoint = pod.Labels["app"]
}
if endpoint == "" {
    return
}

// manager.go - handlePodDelete (重复)
// manager.go - handlePodUpdate (重复)
// jobs.go - podCleanupJob (重复)
```

**优化方案**:
```go
// pkg/deploy/k8s/utils.go
func GetPodEndpoint(pod *corev1.Pod) string {
    if pod == nil || pod.Labels == nil {
        return ""
    }
    return pod.Labels[constants.LabelApp]
}

func IsManagedPod(pod *corev1.Pod) bool {
    return GetPodEndpoint(pod) != ""
}
```

### 3.2 时间格式化 (中优先级)
**问题**: 时间格式化字符串重复

**现状**:
```go
// 多处使用
time.RFC3339
"2006-01-02T15:04:05Z07:00"
```

**优化方案**: 统一使用 `time.RFC3339`

### 3.3 错误处理模式 (低优先级)
**问题**: 错误包装模式不一致

**现状**:
```go
// 有的用 fmt.Errorf
return fmt.Errorf("failed to xxx: %w", err)
// 有的直接返回
return err
```

**优化方案**: 统一使用 `fmt.Errorf` 包装，添加上下文

---

## 四、代码质量优化

### 4.1 manager.go 文件过大 (高优先级)
**问题**: `pkg/deploy/k8s/manager.go` 有 74KB，1400+ 行

**优化方案**: 拆分为多个文件
```
pkg/deploy/k8s/
├── manager.go           # 核心 Manager 结构和初始化
├── callbacks.go         # 所有 callback 注册/通知方法
├── pod_handlers.go      # handlePodAdd/Update/Delete
├── deployment_handlers.go # handleDeploymentEvent/Update
├── informers.go         # Informer 相关
└── utils.go             # 工具函数
```

### 4.2 deployment.go 文件过大 (中优先级)
**问题**: `pkg/interfaces/deployment.go` 包含太多类型定义

**优化方案**: 拆分类型到独立文件

### 4.3 移除废弃注释 (低优先级)
**问题**: 代码中有一些过时的注释

**优化方案**: 清理 `// deprecated`, `// TODO` 等过时注释

---

## 五、配置优化

### 5.1 清理废弃配置 (中优先级)
**问题**: config 中可能有不再使用的配置项

**检查项**:
- Queue 配置是否还需要全部保留
- Providers.Queue 配置是否可以移除

### 5.2 默认值常量化 (低优先级)
**问题**: 默认值散落在代码中

**优化方案**:
```go
// pkg/constants/defaults.go
const (
    DefaultHeartbeatTimeout = 60 * time.Second
    DefaultTaskTimeout      = 300 * time.Second
    DefaultConcurrency      = 1
)
```

---

## 六、执行计划

### Phase 1: 常量定义 (预计 1 小时)
1. 创建 `pkg/constants/` 目录
2. 定义 Worker/Task/Pod/K8s 常量
3. 全局替换硬编码字符串

### Phase 2: 重复代码消除 (预计 30 分钟)
1. 创建 `pkg/deploy/k8s/utils.go`
2. 提取 Pod label 处理逻辑
3. 统一时间格式化

### Phase 3: 目录结构调整 (预计 30 分钟)
1. 移动 `worker_lister.go`
2. 整理 interfaces 目录

### Phase 4: 大文件拆分 (预计 1 小时)
1. 拆分 `manager.go`
2. 拆分 `deployment.go`

### Phase 5: 清理和测试 (预计 30 分钟)
1. 清理废弃代码和注释
2. 编译测试
3. 功能验证

---

## 七、优先级总结

| 优先级 | 任务 | 状态 |
|--------|------|------|
| P0 | Worker/Task 状态常量 | ✅ 完成 |
| P0 | Pod label 提取逻辑 | ✅ 完成 |
| P0 | worker_lister.go 位置 | ✅ 完成 |
| P1 | K8s label 常量 | ✅ 完成 |
| P1 | manager.go 拆分 | 待定（文件较大但功能内聚） |
| P2 | interfaces 目录整理 | ✅ 部分完成 |
| P2 | 配置清理 | 待定 |
| P3 | 其他清理 | 待定 |

## 八、已完成的优化

### 8.1 新增文件
- `pkg/constants/worker.go` - Worker 状态常量
- `pkg/constants/task.go` - Task 状态常量
- `pkg/constants/k8s.go` - K8s label 和 Pod 状态常量
- `pkg/deploy/k8s/utils.go` - K8s 工具函数
- `pkg/interfaces/worker.go` - WorkerLister 接口

### 8.2 重构文件
- `pkg/store/mysql/worker_repository.go` - 使用常量
- `pkg/autoscaler/metrics_collector.go` - 使用常量和接口
- `pkg/autoscaler/executor.go` - 使用接口
- `pkg/autoscaler/manager.go` - 使用接口
- `pkg/autoscaler/worker_lister.go` - 实现接口
- `pkg/deploy/k8s/manager.go` - 使用工具函数和常量
- `internal/service/worker_service.go` - 使用常量
