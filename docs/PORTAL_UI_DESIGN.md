# Portal UI 设计文档

基于现有 Waverless 监控 Dashboard 的设计风格，为 Portal 设计一套完整的用户界面。

## 设计风格特征

### 色彩系统
```css
/* 主色调 */
--primary-blue: #1da1f2;
--primary-hover: #1a91da;

/* 状态色 */
--success-green: #48bb78;
--error-red: #f56565;
--warning-yellow: #ecc94b;
--info-blue: #3b82f6;
--purple: #8b5cf6;
--orange: #f59e0b;

/* 中性色 */
--bg-primary: #f5f7fa;
--bg-white: #ffffff;
--text-primary: #1f2937;
--text-secondary: #6b7280;
--text-muted: #9ca3af;
--border-light: #e1e8ed;
--border-gray: #d1d5db;
```

### 设计原则
1. **简洁现代**：扁平化设计，圆角卡片
2. **数据可视化**：使用 ECharts 图表
3. **响应式布局**：支持桌面和移动端
4. **实时更新**：支持 Live Data 模式
5. **清晰层次**：卡片阴影、边框分隔

---

## 页面结构

### 1. 主导航（顶部）

```
┌─────────────────────────────────────────────────────────────┐
│ [Logo] Portal    Dashboard  Endpoints  Specs  Billing  Docs │
│                                              [User] [Logout] │
└─────────────────────────────────────────────────────────────┘
```

---

## 页面设计

### 1. Dashboard（总览页）

#### 1.1 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ Dashboard                                    [Time Range ▼] │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐│
│ │ Total       │ │ Active      │ │ Total       │ │ Balance ││
│ │ Endpoints   │ │ Workers     │ │ Tasks       │ │         ││
│ │    24       │ │    156      │ │  54,717     │ │ $125.50 ││
│ │ +3 today    │ │ +12 today   │ │ +2.3k today │ │ -$12.30 ││
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Cost Trend (Last 7 Days)                    Total: $86.40││
│ │ [折线图：每日费用趋势]                                    ││
│ └───────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────┐ ┌────────────────────────��┐ │
│ │ Tasks by Status             │ │ Workers by Spec Type    │ │
│ │ [饼图：任务状态分布]        │ │ [饼图：GPU vs CPU]      │ │
│ └─────────────────────────────┘ └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Recent Endpoints                          [View All →]   ││
│ │ ┌─────────────────────────────────────────────────────┐  ││
│ │ │ my-model-1    GPU-A100-40GB  Running  5 workers     │  ││
│ │ │ my-service    CPU-16C-32G    Running  10 workers    │  ││
│ │ │ inference-api GPU-H100-80GB  Running  2 workers     │  ││
│ │ └─────────────────────────────────────────────────────┘  ││
│ └───────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

#### 1.2 关键指标卡片

```html
<div class="metric-card">
  <div class="metric-icon">
    <i class="fas fa-server"></i>
  </div>
  <div class="metric-content">
    <div class="metric-label">Total Endpoints</div>
    <div class="metric-value">24</div>
    <div class="metric-change positive">
      <i class="fas fa-arrow-up"></i> +3 today
    </div>
  </div>
</div>
```

---

### 2. Endpoints（端点管理页）

#### 2.1 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ Endpoints                                   [+ New Endpoint]│
├─────────────────────────────────────────────────────────────┤
│ [Search...] [Filter: All ▼] [Spec Type: All ▼] [Sort ▼]   │
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ my-model-1                                    [⋮ Actions]││
│ │ GPU-A100-40GB • us-east-1 • $2.80/h                      ││
│ │ ● Running  5/10 workers  2,345 tasks  $67.20 (24h)      ││
│ │ [View Details] [Scale] [Stop]                            ││
│ └───────────────────────────────────────────────────────────┘│
│ ┌───────────────────────────────────────────────────────────┐│
│ │ my-service                                    [⋮ Actions]││
│ │ CPU-16C-32G • eu-west-1 • $0.40/h                        ││
│ │ ● Running  10/20 workers  5,678 tasks  $9.60 (24h)      ││
│ │ [View Details] [Scale] [Stop]                            ││
│ └───────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

#### 2.2 Endpoint 详情页

```
┌─────────────────────────────────────────────────────────────┐
│ ← Back to Endpoints                                         │
├─────────────────────────────────────────────────────────────┤
│ my-model-1                                      ● Running   │
│ GPU-A100-40GB • us-east-1 • $2.80/h                        │
├─────────────────────────────────────────────────────────────┤
│ [Overview] [Metrics] [Workers] [Tasks] [Logs] [Settings]   │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐│
│ │ Workers     │ │ Tasks       │ │ Avg Latency │ │ Cost    ││
│ │ 5 / 10      │ │ 2,345       │ │ 125ms       │ │ $67.20  ││
│ │ (50%)       │ │ (24h)       │ │ (P95: 250ms)│ │ (24h)   ││
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Worker Count Trend (Last 24h)                            ││
│ │ [折线图：Worker 数量变化]                                ││
│ └───────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────┐ ┌─────────────────────────┐ │
│ │ Task Throughput             │ │ Cost Breakdown          │ │
│ │ [柱状图：任务吞吐量]        │ │ [面积图：费用趋势]      │ │
│ └─────────────────────────────┘ └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Active Workers                                           ││
│ │ ┌─────────────────────────────────────────────────────┐  ││
│ │ │ worker-abc-123  Running  15 tasks  2.5h  $7.00      │  ││
│ │ │ worker-def-456  Running  12 tasks  1.8h  $5.04      │  ││
│ │ │ worker-ghi-789  Idle     0 tasks   0.3h  $0.84      │  ││
│ │ └─────────────────────────────────────────────────────┘  ││
│ └───────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

### 3. Specs（规格浏览页）

#### 3.1 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ Available Specs                                             │
├─────────────────────────────────────────────────────────────┤
│ [GPU Specs] [CPU Specs] [All]                               │
├─────────────────────────────────────────────────────────────┤
│ GPU Specs                                                   │
│ ┌─────────────────────────────┐ ┌─────────────────────────┐ │
│ │ GPU-A100-40GB               │ │ GPU-H100-80GB           │ │
│ │ NVIDIA A100 40GB GPU        │ │ NVIDIA H100 80GB GPU    │ │
│ │ • 1x A100 GPU               │ │ • 1x H100 GPU           │ │
│ │ • 16 vCPU, 64GB RAM         │ │ • 32 vCPU, 128GB RAM    │ │
│ │ • 200GB Disk                │ │ • 500GB Disk            │ │
│ │                             │ │                         │ │
│ │ $2.80/hour                  │ │ $4.50/hour              │ │
│ │ 3 clusters • 35 available   │ │ 2 clusters • 8 available│ │
│ │ [Create Endpoint]           │ │ [Create Endpoint]       │ │
│ └─────────────────────────────┘ └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ CPU Specs                                                   │
│ ┌─────────────────────────────┐ ┌─────────────────────────┐ │
│ │ CPU-16C-32G                 │ │ CPU-32C-64G             │ │
│ │ 16 vCPU, 32GB RAM           │ │ 32 vCPU, 64GB RAM       │ │
│ │ • 16 vCPU                   │ │ • 32 vCPU               │ │
│ │ • 32GB RAM                  │ │ • 64GB RAM              │ │
│ │ • 200GB Disk                │ │ • 500GB Disk            │ │
│ │                             │ │                         │ │
│ │ $0.40/hour                  │ │ $0.80/hour              │ │
│ │ 5 clusters • 150 available  │ │ 4 clusters • 80 available│
│ │ [Create Endpoint]           │ │ [Create Endpoint]       │ │
│ └─────────────────────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

#### 3.2 创建 Endpoint 对话框

```
┌─────────────────────────────────────────────────────────────┐
│ Create New Endpoint                                    [✕]  │
├─────────────────────────────────────────────────────────────┤
│ Step 1: Basic Information                                   │
│                                                             │
│ Endpoint Name *                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ my-model                                                │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Select Spec *                                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ GPU-A100-40GB ($2.80/h) ▼                               │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Docker Image *                                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ myregistry/my-model:latest                              │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ Step 2: Scaling Configuration                               │
│                                                             │
│ Min Replicas: [1]  Max Replicas: [10]                      │
│                                                             │
│ Task Timeout: [3600] seconds                                │
├─────────────────────────────────────────────────────────────┤
│ Step 3: Environment Variables (Optional)                    │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ KEY=VALUE                                               │ │
│ │ [+ Add Variable]                                        │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ Estimated Cost: $2.80/hour (min) - $28.00/hour (max)       │
│                                                             │
│                              [Cancel] [Create Endpoint]     │
└─────────────────────────────────────────────────────────────┘
```

---

### 4. Billing（计费页）

#### 4.1 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ Billing & Usage                                             │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐│
│ │ Balance     │ │ This Month  │ │ Last Month  │ │ Total   ││
│ │ $125.50     │ │ $342.80     │ │ $298.45     │ │ $1,245  ││
│ │ [Recharge]  │ │ +15.2%      │ │             │ │ (All)   ││
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Cost Trend (Last 30 Days)                  Total: $342.80││
│ │ [面积图：每日费用趋势]                                    ││
│ └───────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────┐ ┌─────────────────────────┐ │
│ │ Cost by Endpoint            │ │ Cost by Spec Type       │ │
│ │ [饼图：各 Endpoint 费用]    │ │ [饼图：GPU vs CPU]      │ │
│ └─────────────────────────────┘ └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Billing History                           [Export CSV]   ││
│ │ [Date Range: Last 30 Days ▼]                             ││
│ │ ┌─────────────────────────────────────────────────────┐  ││
│ │ │ Date       Endpoint      Worker      Hours   Cost   │  ││
│ │ │ 2026-01-05 my-model-1    worker-123  2.5h    $7.00  │  ││
│ │ │ 2026-01-05 my-service    worker-456  5.0h    $2.00  │  ││
│ │ │ 2026-01-04 my-model-1    worker-789  8.0h    $22.40 │  ││
│ │ └─────────────────────────────────────────────────────┘  ││
│ │ [Load More]                                              ││
│ └───────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

#### 4.2 充值对话框

```
┌─────────────────────────────────────────────────────────────┐
│ Recharge Balance                                       [✕]  │
├─────────────────────────────────────────────────────────────┤
│ Current Balance: $125.50                                    │
│                                                             │
│ Select Amount                                               │
│ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐               │
│ │  $50   │ │ $100   │ │ $200   │ │ $500   │               │
│ └────────┘ └────────┘ └────────┘ └────────┘               │
│                                                             │
│ Or enter custom amount:                                     │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ $ 100.00                                                │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Payment Method                                              │
│ ○ Credit Card  ○ PayPal  ○ Stripe                          │
│                                                             │
│ New Balance: $225.50                                        │
│                                                             │
│                              [Cancel] [Proceed to Payment]  │
└─────────────────────────────────────────────────────────────┘
```

---

### 5. Workers（Worker 列表页）

```
┌─────────────────────────────────────────────────────────────┐
│ All Workers                                                 │
├─────────────────────────────────────────────────────────────┤
│ [Search...] [Endpoint: All ▼] [Status: All ▼] [Sort ▼]     │
├─────────────────────────────────────────────────────────────┤
│ ┌───────────────────────────────────────────────────────────┐│
│ │ Worker ID       Endpoint    Status  Tasks  Uptime  Cost  ││
│ │ worker-abc-123  my-model-1  Running 15     2.5h    $7.00 ││
│ │ worker-def-456  my-service  Running 12     1.8h    $0.72 ││
│ │ worker-ghi-789  my-model-1  Idle    0      0.3h    $0.84 ││
│ │ worker-jkl-012  inference   Running 8      4.2h    $11.76││
│ └───────────────────────────────────────────────────────────┘│
│ [Load More]                                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 组件库

### 1. 按钮样式

```css
/* Primary Button */
.btn-primary {
    background: #1da1f2;
    color: white;
    padding: 10px 20px;
    border-radius: 6px;
    border: none;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-primary:hover {
    background: #1a91da;
    box-shadow: 0 2px 8px rgba(29, 161, 242, 0.3);
}

/* Secondary Button */
.btn-secondary {
    background: white;
    color: #1da1f2;
    border: 1px solid #1da1f2;
}

/* Danger Button */
.btn-danger {
    background: #f56565;
    color: white;
}
```

### 2. 卡片样式

```css
.card {
    background: white;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: all 0.2s;
}

.card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
```

### 3. 状态徽章

```css
.badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
}

.badge-success {
    background: #d1fae5;
    color: #065f46;
}

.badge-error {
    background: #fee2e2;
    color: #991b1b;
}

.badge-warning {
    background: #fef3c7;
    color: #92400e;
}
```

---

## 响应式设计

### 断点
- **Desktop**: ≥ 1024px (2列布局)
- **Tablet**: 768px - 1023px (1列布局)
- **Mobile**: < 768px (堆叠布局)

### 移动端适配
```css
@media (max-width: 768px) {
    .charts-grid {
        grid-template-columns: 1fr;
    }

    .metric-cards {
        grid-template-columns: repeat(2, 1fr);
    }

    .nav-tabs {
        overflow-x: auto;
    }
}
```

---

## 交互设计

### 1. 加载状态
- 使用旋转 spinner
- 骨架屏（Skeleton Screen）
- 进度条

### 2. 错误提示
- Toast 通知（右上角）
- 内联错误提示
- 确认对话框

### 3. 实时更新
- WebSocket 连接状态指示
- 自动刷新开关
- 数据更新动画

---

## 技术栈建议

### 前端框架
- **React** + TypeScript
- **Vite** 构建工具
- **TailwindCSS** 样式框架

### 图表库
- **ECharts** (与现有 dashboard 一致)

### 状态管理
- **Zustand** 或 **React Query**

### 路由
- **React Router v6**

### UI 组件库
- **Headless UI** (对话框、下拉菜单)
- **React Icons** (图标)

---

## 实现优先级

### Phase 1 (MVP)
1. Dashboard 总览页
2. Endpoints 列表页
3. Endpoint 详情页
4. Specs 浏览页
5. 创建 Endpoint 功能

### Phase 2
1. Billing 页面
2. Workers 列表页
3. 实时数据更新
4. 导出功能

### Phase 3
1. 高级筛选和搜索
2. 自定义 Dashboard
3. 告警设置
4. 移动端优化

---

## 示例代码片段

### Dashboard 指标卡片组件

```tsx
interface MetricCardProps {
  icon: string;
  label: string;
  value: string | number;
  change?: {
    value: string;
    positive: boolean;
  };
}

export const MetricCard: React.FC<MetricCardProps> = ({
  icon,
  label,
  value,
  change
}) => {
  return (
    <div className="bg-white rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="text-sm text-gray-600 mb-1">{label}</div>
          <div className="text-3xl font-bold text-gray-900">{value}</div>
          {change && (
            <div className={`text-sm mt-2 ${change.positive ? 'text-green-600' : 'text-red-600'}`}>
              <i className={`fas fa-arrow-${change.positive ? 'up' : 'down'} mr-1`}></i>
              {change.value}
            </div>
          )}
        </div>
        <div className="text-3xl text-blue-500">
          <i className={icon}></i>
        </div>
      </div>
    </div>
  );
};
```

### Endpoint 卡片组件

```tsx
interface EndpointCardProps {
  name: string;
  specName: string;
  cluster: string;
  pricePerHour: number;
  status: 'running' | 'stopped' | 'deploying';
  workers: { current: number; max: number };
  tasks24h: number;
  cost24h: number;
}

export const EndpointCard: React.FC<EndpointCardProps> = ({
  name,
  specName,
  cluster,
  pricePerHour,
  status,
  workers,
  tasks24h,
  cost24h
}) => {
  const statusColors = {
    running: 'text-green-600',
    stopped: 'text-gray-600',
    deploying: 'text-yellow-600'
  };

  return (
    <div className="bg-white rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">{name}</h3>
          <div className="text-sm text-gray-600 mt-1">
            {specName} • {cluster} • ${pricePerHour}/h
          </div>
        </div>
        <button className="text-gray-400 hover:text-gray-600">
          <i className="fas fa-ellipsis-v"></i>
        </button>
      </div>

      <div className="flex items-center gap-4 text-sm">
        <span className={`flex items-center gap-1 ${statusColors[status]}`}>
          <i className="fas fa-circle text-xs"></i>
          {status.charAt(0).toUpperCase() + status.slice(1)}
        </span>
        <span className="text-gray-600">
          {workers.current}/{workers.max} workers
        </span>
        <span className="text-gray-600">
          {tasks24h.toLocaleString()} tasks
        </span>
        <span className="text-gray-900 font-semibold">
          ${cost24h.toFixed(2)} (24h)
        </span>
      </div>

      <div className="flex gap-2 mt-4">
        <button className="btn-primary flex-1">View Details</button>
        <button className="btn-secondary">Scale</button>
        <button className="btn-secondary">Stop</button>
      </div>
    </div>
  );
};
```

---

## 总结

这套 UI 设计：
1. ✅ **风格统一**：与现有 Waverless dashboard 保持一致
2. ✅ **功能完整**：覆盖 Portal 所有核心功能
3. ✅ **用户友好**：清晰的信息层次和交互流程
4. ✅ **可扩展**：模块化设计，易于添加新功能
5. ✅ **响应式**：支持桌面和移动端

建议使用 React + TypeScript + TailwindCSS + ECharts 技术栈实现。
