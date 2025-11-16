# Spec Management System - 完整实现指南

## 概述

已实现完整的基于数据库的规格（Spec）管理系统，支持运行时动态CRUD操作，无需重启服务。

## 实现的功能

### 1. 数据库层
- ✅ `resource_specs` 表 - 存储规格配置
- ✅ SpecRepository - 数据访问层
- ✅ 支持按类别查询、状态管理

### 2. 后端API
- ✅ `POST /api/v1/k8s/specs` - 创建规格
- ✅ `GET /api/v1/k8s/specs` - 列出规格（支持按category过滤）
- ✅ `GET /api/v1/k8s/specs/:name` - 获取规格详情
- ✅ `PUT /api/v1/k8s/specs/:name` - 更新规格
- ✅ `DELETE /api/v1/k8s/specs/:name` - 删除规格（软删除）

### 3. 核心机制
- ✅ SpecManager优先从数据库读取
- ✅ 数据库不可用时回退到YAML配置
- ✅ 运行时动态更新，无需重启

### 4. Web UI
- ✅ 规格列表展示
- ✅ 创建规格表单
- ✅ 编辑规格表单
- ✅ 删除规格（带确认）
- ✅ 按类别过滤

## 部署步骤

### 1. 数据库迁移

```bash
# 执行SQL脚本创建resource_specs表
mysql -u root -p waverless < scripts/init.sql

# 或者只执行resource_specs表的创建语句
mysql -u root -p waverless -e "
CREATE TABLE \`resource_specs\` (
  \`id\` bigint NOT NULL AUTO_INCREMENT,
  \`name\` varchar(100) NOT NULL COMMENT 'Spec name (unique identifier)',
  \`display_name\` varchar(255) NOT NULL COMMENT 'Display name',
  \`category\` varchar(50) NOT NULL COMMENT 'Category: cpu, gpu',
  \`cpu\` varchar(50) DEFAULT NULL COMMENT 'CPU cores',
  \`memory\` varchar(50) NOT NULL COMMENT 'Memory',
  \`gpu\` varchar(50) DEFAULT NULL COMMENT 'GPU count',
  \`gpu_type\` varchar(100) DEFAULT NULL COMMENT 'GPU type',
  \`ephemeral_storage\` varchar(50) NOT NULL COMMENT 'Ephemeral storage',
  \`platforms\` json DEFAULT NULL COMMENT 'Platform-specific configurations',
  \`status\` varchar(50) NOT NULL DEFAULT 'active' COMMENT 'Spec status',
  \`created_at\` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  \`updated_at\` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (\`id\`),
  UNIQUE KEY \`idx_spec_name_unique\` (\`name\`),
  KEY \`idx_category\` (\`category\`),
  KEY \`idx_status\` (\`status\`),
  KEY \`idx_created_at\` (\`created_at\`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
"
```

### 2. 编译后端

```bash
cd /Users/shanliu/work/wavespeedai/waverless

# 编译
go build -o waverless cmd/*.go

# 或使用现有的构建脚本
./build-and-push.sh
```

### 3. 编译前端

```bash
cd /Users/shanliu/work/wavespeedai/waverless/web-ui

# 安装依赖（如果需要）
pnpm install

# 构建
pnpm run build
```

### 4. 启动服务

```bash
# 启动后端
./waverless

# 服务会在初始化时输出：
# "Spec service injected into SpecManager - specs will be read from database first"
```

## 测试步骤

### 1. API测试

```bash
# 1. 创建GPU规格
curl -X POST http://localhost:8080/api/v1/k8s/specs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "h200-single",
    "displayName": "H200 1 GPU",
    "category": "gpu",
    "resources": {
      "cpu": "16",
      "memory": "64Gi",
      "gpu": "1",
      "gpuType": "NVIDIA-H200",
      "ephemeralStorage": "300"
    },
    "platforms": {
      "generic": {
        "nodeSelector": {
          "gpu.nvidia.com/class": "H200"
        },
        "tolerations": [],
        "labels": {},
        "annotations": {}
      }
    }
  }'

# 2. 创建CPU规格
curl -X POST http://localhost:8080/api/v1/k8s/specs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cpu-2c4g",
    "displayName": "CPU 2 Cores 4GB",
    "category": "cpu",
    "resources": {
      "cpu": "2",
      "memory": "4Gi",
      "ephemeralStorage": "30"
    },
    "platforms": {
      "generic": {
        "nodeSelector": {},
        "tolerations": [],
        "labels": {},
        "annotations": {}
      }
    }
  }'

# 3. 列出所有规格
curl http://localhost:8080/api/v1/k8s/specs

# 4. 按类别过滤
curl http://localhost:8080/api/v1/k8s/specs?category=gpu

# 5. 获取单个规格
curl http://localhost:8080/api/v1/k8s/specs/h200-single

# 6. 更新规格
curl -X PUT http://localhost:8080/api/v1/k8s/specs/h200-single \
  -H "Content-Type: application/json" \
  -d '{
    "displayName": "H200 Single GPU (Updated)"
  }'

# 7. 删除规格
curl -X DELETE http://localhost:8080/api/v1/k8s/specs/cpu-2c4g
```

### 2. Web UI测试

1. 访问 http://localhost:8080 (或你的服务地址)
2. 导航到 "Specs" 页面
3. 测试以下功能：
   - ✅ 查看规格列表
   - ✅ 点击 "Create Spec" 创建新规格
   - ✅ 使用类别过滤器（ALL / CPU / GPU）
   - ✅ 点击编辑图标修改规格
   - ✅ 点击删除图标删除规格（带确认）

### 3. 验证动态更新

```bash
# 1. 创建一个新的spec
curl -X POST http://localhost:8080/api/v1/k8s/specs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-spec",
    "displayName": "Test Spec",
    "category": "cpu",
    "resources": {
      "cpu": "1",
      "memory": "2Gi",
      "ephemeralStorage": "20"
    }
  }'

# 2. 立即使用这个spec创建endpoint（无需重启服务）
curl -X POST http://localhost:8080/api/v1/endpoints \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "test-app",
    "specName": "test-spec",
    "image": "your-image:latest",
    "replicas": 1
  }'
```

## 数据迁移（可选）

如果需要将现有的 `config/specs.yaml` 迁移到数据库：

```bash
# 1. 读取specs.yaml文件
# 2. 对每个spec调用创建API

# 示例脚本：
cat config/specs.yaml | yq '.specs[] | {
  name: .name,
  displayName: .displayName,
  category: .category,
  resources: .resources,
  platforms: .platforms
}' | while read spec; do
  curl -X POST http://localhost:8080/api/v1/k8s/specs \
    -H "Content-Type: application/json" \
    -d "$spec"
done
```

## 架构说明

### 数据流

```
Web UI (React)
    ↓
API Handler (spec_handler.go)
    ↓
SpecService (spec_service.go)
    ↓
SpecRepository (spec_repository.go)
    ↓
MySQL (resource_specs表)
    ↑
SpecManager (spec.go) - 优先读取数据库，回退到YAML
    ↑
K8sDeploymentProvider - 使用spec创建部署
```

### 优先级机制

1. **数据库优先**：SpecManager的`GetSpec()`和`ListSpecs()`方法优先从数据库读取
2. **YAML回退**：如果数据库不可用或没有数据，回退到specs.yaml
3. **无缝切换**：无需修改现有代码，完全向后兼容

## 故障排查

### 问题1：创建spec失败
```bash
# 检查日志
tail -f logs/waverless.log | grep "spec"

# 常见原因：
# - name重复（unique key冲突）
# - 必填字段缺失
# - 数据库连接失败
```

### 问题2：spec没有从数据库读取
```bash
# 检查初始化日志，应该看到：
# "Spec service injected into SpecManager - specs will be read from database first"

# 如果没有这条日志，检查：
# - 数据库配置是否正确
# - spec表是否创建成功
```

### 问题3：Web UI无法显示specs
```bash
# 检查浏览器控制台
# 检查网络请求：/api/v1/k8s/specs

# 检查后端日志
grep "k8s/specs" logs/waverless.log
```

## 最佳实践

1. **初始数据导入**：首次部署时，将常用的spec导入数据库
2. **备份**：定期备份resource_specs表
3. **命名规范**：使用清晰的spec名称，如 `h200-single`、`cpu-4c8g`
4. **分类管理**：合理使用category字段（cpu/gpu）
5. **平台配置**：为不同的K8s平台配置不同的platforms

## 后续优化建议

1. **批量导入**：添加批量导入spec的API
2. **版本控制**：记录spec的修改历史
3. **权限管理**：添加spec管理的权限控制
4. **校验增强**：添加更严格的资源配置校验
5. **模板功能**：提供spec模板快速创建

## 支持

如有问题，请查看：
- 后端日志：`logs/waverless.log`
- 数据库日志：MySQL错误日志
- Web控制台：浏览器开发者工具
