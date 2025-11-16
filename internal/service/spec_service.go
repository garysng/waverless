package service

import (
	"context"
	"fmt"

	"waverless/pkg/interfaces"
	"waverless/pkg/store/mysql"
	"waverless/pkg/store/mysql/model"
)

// SpecService handles spec business logic
type SpecService struct {
	specRepo *mysql.SpecRepository
}

// NewSpecService creates a new spec service
func NewSpecService(specRepo *mysql.SpecRepository) *SpecService {
	return &SpecService{
		specRepo: specRepo,
	}
}

// CreateSpec creates a new spec
func (s *SpecService) CreateSpec(ctx context.Context, req *interfaces.CreateSpecRequest) (*interfaces.SpecInfo, error) {
	// Check if spec with same name already exists
	existing, err := s.specRepo.Get(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing spec: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("spec with name %s already exists", req.Name)
	}

	// Create spec model
	spec := &model.Spec{
		Name:             req.Name,
		DisplayName:      req.DisplayName,
		Category:         req.Category,
		ResourceType:     req.ResourceType,
		CPU:              req.Resources.CPU,
		Memory:           req.Resources.Memory,
		GPU:              req.Resources.GPU,
		GPUType:          req.Resources.GPUType,
		EphemeralStorage: req.Resources.EphemeralStorage,
		ShmSize:          req.Resources.ShmSize,
		Platforms:        req.Platforms,
		Status:           "active",
	}

	if err := s.specRepo.Create(ctx, spec); err != nil {
		return nil, fmt.Errorf("failed to create spec: %w", err)
	}

	return s.modelToSpecInfo(spec), nil
}

// GetSpec retrieves a spec by name
func (s *SpecService) GetSpec(ctx context.Context, name string) (*interfaces.SpecInfo, error) {
	spec, err := s.specRepo.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if spec == nil {
		return nil, fmt.Errorf("spec not found: %s", name)
	}
	return s.modelToSpecInfo(spec), nil
}

// ListSpecs lists all active specs
func (s *SpecService) ListSpecs(ctx context.Context) ([]*interfaces.SpecInfo, error) {
	specs, err := s.specRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*interfaces.SpecInfo, len(specs))
	for i, spec := range specs {
		result[i] = s.modelToSpecInfo(spec)
	}
	return result, nil
}

// ListSpecsByCategory lists specs by category
func (s *SpecService) ListSpecsByCategory(ctx context.Context, category string) ([]*interfaces.SpecInfo, error) {
	specs, err := s.specRepo.ListByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	result := make([]*interfaces.SpecInfo, len(specs))
	for i, spec := range specs {
		result[i] = s.modelToSpecInfo(spec)
	}
	return result, nil
}

// UpdateSpec updates a spec
func (s *SpecService) UpdateSpec(ctx context.Context, name string, req *interfaces.UpdateSpecRequest) (*interfaces.SpecInfo, error) {
	spec, err := s.specRepo.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if spec == nil {
		return nil, fmt.Errorf("spec not found: %s", name)
	}

	// Update fields if provided
	if req.DisplayName != nil {
		spec.DisplayName = *req.DisplayName
	}
	if req.Category != nil {
		spec.Category = *req.Category
	}
	if req.ResourceType != nil {
		spec.ResourceType = *req.ResourceType
	}
	if req.Resources != nil {
		if req.Resources.CPU != "" {
			spec.CPU = req.Resources.CPU
		}
		if req.Resources.Memory != "" {
			spec.Memory = req.Resources.Memory
		}
		if req.Resources.GPU != "" {
			spec.GPU = req.Resources.GPU
		}
		if req.Resources.GPUType != "" {
			spec.GPUType = req.Resources.GPUType
		}
		if req.Resources.EphemeralStorage != "" {
			spec.EphemeralStorage = req.Resources.EphemeralStorage
		}
		if req.Resources.ShmSize != "" {
			spec.ShmSize = req.Resources.ShmSize
		}
	}
	if req.Platforms != nil {
		spec.Platforms = req.Platforms
	}
	if req.Status != nil {
		spec.Status = *req.Status
	}

	if err := s.specRepo.Update(ctx, spec); err != nil {
		return nil, fmt.Errorf("failed to update spec: %w", err)
	}

	return s.modelToSpecInfo(spec), nil
}

// DeleteSpec deletes a spec
func (s *SpecService) DeleteSpec(ctx context.Context, name string) error {
	return s.specRepo.Delete(ctx, name)
}

// modelToSpecInfo converts model.Spec to interfaces.SpecInfo
func (s *SpecService) modelToSpecInfo(spec *model.Spec) *interfaces.SpecInfo {
	return &interfaces.SpecInfo{
		Name:         spec.Name,
		DisplayName:  spec.DisplayName,
		Category:     spec.Category,
		ResourceType: spec.ResourceType,
		Resources: interfaces.ResourceRequirements{
			CPU:              spec.CPU,
			Memory:           spec.Memory,
			GPU:              spec.GPU,
			GPUType:          spec.GPUType,
			EphemeralStorage: spec.EphemeralStorage,
			ShmSize:          spec.ShmSize,
		},
		Platforms: spec.Platforms,
	}
}
