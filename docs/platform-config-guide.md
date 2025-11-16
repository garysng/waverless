# Platform Configuration Guide

## Overview

The Spec Management system now includes a powerful **Platform Configuration Editor** that allows you to manage complex Kubernetes configurations (tolerations, node selectors, labels, annotations) through a user-friendly interface.

## EphemeralStorage

**ephemeralStorage**: Temporary storage size (EmptyDir, container writable layer)
- Data is **lost** when Pod is deleted
- Includes container writable layer + emptyDir volumes
- Kubernetes resource limit field: `resources.limits.ephemeral-storage`

### Future Enhancement

The deployment template ([config/templates/deployment.yaml](../config/templates/deployment.yaml)) currently does **not** use this field. Recommended addition:

```yaml
resources:
  requests:
    memory: "{{.MemoryRequest}}"
    cpu: "{{.CpuLimit}}"
    ephemeral-storage: "{{.EphemeralStorage}}"
  limits:
    memory: "{{.MemoryRequest}}"
    cpu: "{{.CpuLimit}}"
    ephemeral-storage: "{{.EphemeralStorage}}"
```

### Persistent Storage (PVC)

For persistent storage needs, PVC (PersistentVolumeClaim) will be supported in a future version with a separate configuration system.

## Platform Configuration Editor

### Features

The Platform Configuration Editor supports **three modes**:

1. **Simple Mode (Form-based)**: User-friendly interface for managing configurations
2. **YAML Mode**: YAML-based editing with automatic JSON conversion
3. **JSON Mode (Advanced)**: Direct JSON editing for power users

### Simple Mode Usage

#### 1. Add a Platform

Click **"Add Platform"** and enter platform name (e.g., `generic`, `aliyun-ack`, `aws-eks`)

#### 2. Configure Tolerations

Tolerations allow Pods to schedule on nodes with specific taints.

**Example**: Schedule on H200 GPU nodes
```yaml
tolerations:
  - key: "hardware-type/h200"
    operator: "Equal"
    value: "gpu"
    effect: "NoSchedule"
```

**Steps**:
1. Expand the "Tolerations" section
2. Click "Add Toleration"
3. Fill in:
   - **Key**: `hardware-type/h200`
   - **Operator**: `Equal` or `Exists`
   - **Value**: `gpu` (only for `Equal` operator)
   - **Effect**: `NoSchedule`, `PreferNoSchedule`, or `NoExecute`

#### 3. Configure Node Selectors

Node selectors restrict Pod scheduling to nodes with specific labels.

**Example**: Select GPU nodes
```yaml
nodeSelector:
  gpu.nvidia.com/class: "H200"
```

**Steps**:
1. Expand the "Node Selector" section
2. Click "Add Node Selector" - a new inline form card appears
3. Enter key in the first input field: `gpu.nvidia.com/class`
4. Enter value in the second input field: `H200`
5. Edit inline or click the delete button to remove

#### 4. Configure Labels

Add custom labels to Pods.

**Example**: Aliyun ACK labels
```yaml
labels:
  alibabacloud.com/acs: "true"
  alibabacloud.com/compute-class: "general-purpose"
```

**Steps**:
1. Expand the "Labels" section
2. Click "Add Label" - a new inline form card appears
3. Enter key in the first input field (e.g., `alibabacloud.com/acs`)
4. Enter value in the second input field (e.g., `true`)
5. Edit inline or click the delete button to remove

#### 5. Configure Annotations

Add custom annotations to Pods.

**Example**: Enable Aliyun image acceleration
```yaml
annotations:
  k8s.aliyun.com/image-accelerate-mode: "on-demand"
```

**Steps**:
1. Expand the "Annotations" section
2. Click "Add Annotation" - a new inline form card appears
3. Enter key in the first input field (e.g., `k8s.aliyun.com/image-accelerate-mode`)
4. Enter value in the second input field (e.g., `on-demand`)
5. Edit inline or click the delete button to remove

### YAML Mode Usage

Switch to **"YAML Mode"** tab for YAML-based editing with automatic JSON conversion.

**Example**: Complete platform configuration in YAML
```yaml
generic:
  nodeSelector: {}
  tolerations:
    - key: hardware-type/h200
      operator: Equal
      value: gpu
      effect: NoSchedule
  labels: {}
  annotations: {}
aliyun-ack:
  nodeSelector:
    gpu.nvidia.com/class: H200
  tolerations:
    - key: hardware-type/h200
      operator: Equal
      value: gpu-cpfs
      effect: NoSchedule
  labels:
    alibabacloud.com/acs: "true"
    alibabacloud.com/compute-class: general-purpose
  annotations:
    k8s.aliyun.com/image-accelerate-mode: on-demand
```

**Features**:
- **Real-time validation**: Displays errors if YAML is invalid
- **Copy YAML**: Copy the YAML to clipboard
- **Format YAML**: Auto-format the YAML with proper indentation
- **Convert to JSON**: Convert YAML to JSON and copy to clipboard
- **Automatic conversion**: YAML is automatically converted to JSON when submitting the form

**Advantages**:
- More readable than JSON for complex configurations
- No need to escape quotes or manage commas
- Supports comments (which are stripped during conversion)
- Familiar format for Kubernetes users

### JSON Mode Usage

Switch to **"JSON Mode"** tab for direct JSON editing.

**Example**: Complete platform configuration
```json
{
  "generic": {
    "nodeSelector": {},
    "tolerations": [
      {
        "key": "hardware-type/h200",
        "operator": "Equal",
        "value": "gpu",
        "effect": "NoSchedule"
      }
    ],
    "labels": {},
    "annotations": {}
  },
  "aliyun-ack": {
    "nodeSelector": {
      "gpu.nvidia.com/class": "H200"
    },
    "tolerations": [
      {
        "key": "hardware-type/h200",
        "operator": "Equal",
        "value": "gpu-cpfs",
        "effect": "NoSchedule"
      }
    ],
    "labels": {
      "alibabacloud.com/acs": "true",
      "alibabacloud.com/compute-class": "general-purpose",
      "alibabacloud.com/compute-qos": "default"
    },
    "annotations": {
      "k8s.aliyun.com/image-accelerate-mode": "on-demand"
    }
  }
}
```

**Features**:
- **Copy JSON**: Copy the JSON to clipboard
- **Format JSON**: Auto-format the JSON with proper indentation
- **Real-time validation**: Displays errors if JSON is invalid

## Complete Example: Creating H200 GPU Spec

### Via Web UI (Simple Mode)

1. Navigate to **Specs** page
2. Click **"Create Spec"**
3. Fill in basic information:
   - **Spec Name**: `h200-single`
   - **Display Name**: `H200 1 GPU`
   - **Category**: `gpu`
   - **CPU Cores**: `16`
   - **Memory**: `64Gi`
   - **GPU Count**: `1`
   - **GPU Type**: `NVIDIA-H200`
   - **Ephemeral Storage**: `300`

4. Configure platforms:
   - Click **"Add Platform"** → Enter `aliyun-ack`
   - Add toleration:
     - Key: `hardware-type/h200`
     - Operator: `Equal`
     - Value: `gpu-cpfs`
     - Effect: `NoSchedule`
   - Add node selector:
     - Key: `gpu.nvidia.com/class`
     - Value: `H200`
   - Add annotation:
     - Key: `k8s.aliyun.com/image-accelerate-mode`
     - Value: `on-demand`

5. Click **"Create"**

### Via API

```bash
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
      "aliyun-ack": {
        "nodeSelector": {
          "gpu.nvidia.com/class": "H200"
        },
        "tolerations": [
          {
            "key": "hardware-type/h200",
            "operator": "Equal",
            "value": "gpu-cpfs",
            "effect": "NoSchedule"
          }
        ],
        "labels": {
          "alibabacloud.com/acs": "true"
        },
        "annotations": {
          "k8s.aliyun.com/image-accelerate-mode": "on-demand"
        }
      }
    }
  }'
```

## Common Platform Configurations

### Generic (Default)

```json
{
  "generic": {
    "nodeSelector": {},
    "tolerations": [],
    "labels": {},
    "annotations": {}
  }
}
```

### Aliyun ACK with GPU

```json
{
  "aliyun-ack": {
    "nodeSelector": {
      "gpu.nvidia.com/class": "H200"
    },
    "tolerations": [
      {
        "key": "hardware-type/h200",
        "operator": "Equal",
        "value": "gpu-cpfs",
        "effect": "NoSchedule"
      }
    ],
    "annotations": {
      "k8s.aliyun.com/image-accelerate-mode": "on-demand"
    }
  }
}
```

### AWS EKS with GPU

```json
{
  "aws-eks": {
    "nodeSelector": {
      "node.kubernetes.io/instance-type": "p4d.24xlarge"
    },
    "tolerations": [
      {
        "key": "nvidia.com/gpu",
        "operator": "Exists",
        "effect": "NoSchedule"
      }
    ],
    "labels": {
      "workload-type": "gpu-intensive"
    }
  }
}
```

## Best Practices

1. **Platform Naming**:
   - Use descriptive names: `generic`, `aliyun-ack`, `aws-eks`, `gcp-gke`
   - Be consistent across specs

2. **Tolerations**:
   - Use `Equal` operator when you need specific values
   - Use `Exists` operator for general GPU access
   - Choose appropriate effects:
     - `NoSchedule`: Hard requirement
     - `PreferNoSchedule`: Soft preference
     - `NoExecute`: Evict existing Pods

3. **Node Selectors**:
   - Use specific labels for GPU types
   - Avoid overly restrictive selectors

4. **Annotations**:
   - Use cloud-provider specific annotations for optimizations
   - Document custom annotations

5. **Mode Selection**:
   - Use **Simple Mode** for:
     - Standard configurations
     - Step-by-step guided setup
     - Adding one or two key-value pairs
   - Use **YAML Mode** for:
     - Kubernetes-native YAML format
     - Copy-paste from K8s documentation
     - Multi-platform configurations
     - More readable complex configs
   - Use **JSON Mode** for:
     - Programmatic generation
     - API integration
     - When strict JSON formatting is needed

## Troubleshooting

### Platform configuration not applied

**Check**: Verify the platform name matches your K8s platform setting in `config.yaml`:
```yaml
k8s:
  platform: aliyun-ack  # Must match platform name in spec
```

### Pods not scheduling

**Check tolerations**: Ensure tolerations match node taints:
```bash
# List node taints
kubectl describe nodes | grep Taints

# Verify toleration in spec matches node taint
```

**Check node selectors**: Ensure nodes have the required labels:
```bash
# List node labels
kubectl get nodes --show-labels

# Check specific label
kubectl get nodes -l gpu.nvidia.com/class=H200
```

### YAML validation errors

**Use Format YAML**: Click "Format YAML" to auto-fix indentation
**Check syntax**: Look for:
- Incorrect indentation (use 2 spaces, not tabs)
- Missing colons after keys
- Improper list formatting (should use `-` prefix)
- Mixing tabs and spaces

**Common fixes**:
```yaml
# Wrong - using tabs
nodeSelector:
	gpu.nvidia.com/class: H200

# Correct - using spaces
nodeSelector:
  gpu.nvidia.com/class: H200

# Wrong - missing dash for list items
tolerations:
  key: nvidia.com/gpu
  operator: Exists

# Correct - proper list format
tolerations:
  - key: nvidia.com/gpu
    operator: Exists
```

### JSON validation errors

**Use Format JSON**: Click "Format JSON" to auto-fix indentation
**Check syntax**: Look for:
- Missing commas
- Unmatched brackets
- Invalid JSON values (use strings for all values)
- Trailing commas (not allowed in JSON)

## Reference

- [Kubernetes Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
- [Kubernetes Node Selectors](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector)
- [Kubernetes Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [Kubernetes Annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
