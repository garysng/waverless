package k8s

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"waverless/pkg/interfaces"
	"waverless/pkg/logger"
)

// ResourceSpec 资源规格定义
type ResourceSpec struct {
	Name         string                       `yaml:"name" json:"name"`
	DisplayName  string                       `yaml:"displayName" json:"displayName"`
	Category     string                       `yaml:"category" json:"category"` // cpu, gpu
	Resources    SpecResources                `yaml:"resources" json:"resources"`
	Platforms    map[string]PlatformConfig    `yaml:"platforms" json:"platforms"`
}

// SpecResources 规格资源
type SpecResources struct {
	CPU               string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory            string `yaml:"memory" json:"memory"`
	GPU               string `yaml:"gpu,omitempty" json:"gpu,omitempty"`
	GpuType           string `yaml:"gpuType,omitempty" json:"gpuType,omitempty"`
	EphemeralStorage  string `yaml:"ephemeralStorage" json:"ephemeralStorage"`
	ShmSize           string `yaml:"shmSize,omitempty" json:"shmSize,omitempty"` // Shared memory size
}

// PlatformConfig 平台特定配置
type PlatformConfig struct {
	NodeSelector map[string]string `yaml:"nodeSelector" json:"nodeSelector"`
	Tolerations  []Toleration      `yaml:"tolerations" json:"tolerations"`
	Labels       map[string]string `yaml:"labels" json:"labels"`
	Annotations  map[string]string `yaml:"annotations" json:"annotations"`
}

// Toleration 容忍度
type Toleration struct {
	Key      string `yaml:"key" json:"key"`
	Operator string `yaml:"operator" json:"operator"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`
	Effect   string `yaml:"effect" json:"effect"`
}

// SpecsConfig 规格配置文件
type SpecsConfig struct {
	Specs []ResourceSpec `yaml:"specs"`
}

// SpecRepositoryInterface defines the interface for spec repository
type SpecRepositoryInterface interface {
	GetSpec(ctx context.Context, name string) (*interfaces.SpecInfo, error)
	ListSpecs(ctx context.Context) ([]*interfaces.SpecInfo, error)
	ListSpecsByCategory(ctx context.Context, category string) ([]*interfaces.SpecInfo, error)
}

// SpecManager 规格管理器
type SpecManager struct {
	specs       map[string]*ResourceSpec
	specRepo    SpecRepositoryInterface // Database repository (optional, takes priority if available)
}

// NewSpecManager 创建规格管理器
func NewSpecManager(configPath string) (*SpecManager, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs config: %v", err)
	}

	var config SpecsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse specs config: %v", err)
	}

	specs := make(map[string]*ResourceSpec)
	for i := range config.Specs {
		spec := &config.Specs[i]
		specs[spec.Name] = spec
	}

	return &SpecManager{
		specs: specs,
	}, nil
}

// SetSpecRepository sets the spec repository for database access
func (m *SpecManager) SetSpecRepository(repo SpecRepositoryInterface) {
	m.specRepo = repo
}

// GetSpec 获取规格 (优先从数据库读取，如果数据库不可用则从内存读取)
func (m *SpecManager) GetSpec(name string) (*ResourceSpec, error) {
	// Try database first if repository is available
	if m.specRepo != nil {
		dbSpec, err := m.specRepo.GetSpec(context.Background(), name)
		if err == nil && dbSpec != nil {
			// Convert database spec to k8s ResourceSpec
			return m.convertSpecInfoToResourceSpec(dbSpec), nil
		}
		logger.WarnCtx(context.Background(), "Failed to get spec from database, falling back to YAML: %v", err)
	}

	// Fallback to YAML-based specs
	spec, exists := m.specs[name]
	if !exists {
		return nil, fmt.Errorf("spec not found: %s", name)
	}
	return spec, nil
}

// ListSpecs 列出所有规格 (优先从数据库读取)
func (m *SpecManager) ListSpecs() []*ResourceSpec {
	// Try database first if repository is available
	if m.specRepo != nil {
		dbSpecs, err := m.specRepo.ListSpecs(context.Background())
		if err == nil && len(dbSpecs) > 0 {
			result := make([]*ResourceSpec, len(dbSpecs))
			for i, spec := range dbSpecs {
				result[i] = m.convertSpecInfoToResourceSpec(spec)
			}
			return result
		}
		logger.WarnCtx(context.Background(), "Failed to list specs from database, falling back to YAML: %v", err)
	}

	// Fallback to YAML-based specs
	result := make([]*ResourceSpec, 0, len(m.specs))
	for _, spec := range m.specs {
		result = append(result, spec)
	}
	return result
}

// ListSpecsByCategory 按类别列出规格 (优先从数据库读取)
func (m *SpecManager) ListSpecsByCategory(category string) []*ResourceSpec {
	// Try database first if repository is available
	if m.specRepo != nil {
		dbSpecs, err := m.specRepo.ListSpecsByCategory(context.Background(), category)
		if err == nil && len(dbSpecs) > 0 {
			result := make([]*ResourceSpec, len(dbSpecs))
			for i, spec := range dbSpecs {
				result[i] = m.convertSpecInfoToResourceSpec(spec)
			}
			return result
		}
		logger.WarnCtx(context.Background(), "Failed to list specs from database, falling back to YAML: %v", err)
	}

	// Fallback to YAML-based specs
	result := make([]*ResourceSpec, 0)
	for _, spec := range m.specs {
		if spec.Category == category {
			result = append(result, spec)
		}
	}
	return result
}

// GetPlatformConfig 获取平台特定配置
func (s *ResourceSpec) GetPlatformConfig(platform string) PlatformConfig {
	if config, exists := s.Platforms[platform]; exists {
		return config
	}
	// Return generic config as default
	if config, exists := s.Platforms["generic"]; exists {
		return config
	}
	return PlatformConfig{}
}

// convertSpecInfoToResourceSpec converts interfaces.SpecInfo to k8s.ResourceSpec
func (m *SpecManager) convertSpecInfoToResourceSpec(specInfo *interfaces.SpecInfo) *ResourceSpec {
	ctx := context.Background()
	// Convert platforms from map[string]interface{} to map[string]PlatformConfig
	platforms := make(map[string]PlatformConfig)
	if specInfo.Platforms != nil {
		logger.InfoCtx(ctx, "[SPEC-CONVERT] specInfo.Platforms=%+v", specInfo.Platforms)
		for platformName, platformData := range specInfo.Platforms {
			logger.InfoCtx(ctx, "[SPEC-CONVERT] platformName=%s, platformData type=%T, value=%+v", platformName, platformData, platformData)
			if platformMap, ok := platformData.(map[string]interface{}); ok {
				platform := PlatformConfig{}

				// Convert nodeSelector
				logger.InfoCtx(ctx, "[SPEC-CONVERT] platformMap[nodeSelector] type=%T, value=%+v", platformMap["nodeSelector"], platformMap["nodeSelector"])
				if nodeSelector, ok := platformMap["nodeSelector"].(map[string]interface{}); ok {
					platform.NodeSelector = make(map[string]string)
					for k, v := range nodeSelector {
						if str, ok := v.(string); ok {
							platform.NodeSelector[k] = str
						}
					}
					logger.InfoCtx(ctx, "[SPEC-CONVERT] converted NodeSelector=%+v", platform.NodeSelector)
				}

				// Convert labels
				if labels, ok := platformMap["labels"].(map[string]interface{}); ok {
					platform.Labels = make(map[string]string)
					for k, v := range labels {
						if str, ok := v.(string); ok {
							platform.Labels[k] = str
						}
					}
				}

				// Convert annotations
				if annotations, ok := platformMap["annotations"].(map[string]interface{}); ok {
					platform.Annotations = make(map[string]string)
					for k, v := range annotations {
						if str, ok := v.(string); ok {
							platform.Annotations[k] = str
						}
					}
				}

				// Convert tolerations
				if tolerationsData, ok := platformMap["tolerations"].([]interface{}); ok {
					platform.Tolerations = make([]Toleration, 0, len(tolerationsData))
					for _, t := range tolerationsData {
						if tolMap, ok := t.(map[string]interface{}); ok {
							toleration := Toleration{}
							if key, ok := tolMap["key"].(string); ok {
								toleration.Key = key
							}
							if operator, ok := tolMap["operator"].(string); ok {
								toleration.Operator = operator
							}
							if value, ok := tolMap["value"].(string); ok {
								toleration.Value = value
							}
							if effect, ok := tolMap["effect"].(string); ok {
								toleration.Effect = effect
							}
							platform.Tolerations = append(platform.Tolerations, toleration)
						}
					}
				}

				platforms[platformName] = platform
			}
		}
	}

	return &ResourceSpec{
		Name:        specInfo.Name,
		DisplayName: specInfo.DisplayName,
		Category:    specInfo.Category,
		Resources: SpecResources{
			CPU:              specInfo.Resources.CPU,
			Memory:           specInfo.Resources.Memory,
			GPU:              specInfo.Resources.GPU,
			GpuType:          specInfo.Resources.GPUType,
			EphemeralStorage: specInfo.Resources.EphemeralStorage,
			ShmSize:          specInfo.Resources.ShmSize,
		},
		Platforms: platforms,
	}
}
