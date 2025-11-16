package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"waverless/pkg/store/mysql/model"
)

// SpecRepository handles spec persistence in MySQL
type SpecRepository struct {
	ds *Datastore
}

// NewSpecRepository creates a new spec repository
func NewSpecRepository(ds *Datastore) *SpecRepository {
	return &SpecRepository{ds: ds}
}

// Create creates a new spec
func (r *SpecRepository) Create(ctx context.Context, spec *model.Spec) error {
	return r.ds.DB(ctx).Create(spec).Error
}

// Get retrieves a spec by name
func (r *SpecRepository) Get(ctx context.Context, name string) (*model.Spec, error) {
	var spec model.Spec
	err := r.ds.DB(ctx).Where("name = ? AND status != ?", name, "deleted").First(&spec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get spec: %w", err)
	}
	return &spec, nil
}

// List lists all active specs
func (r *SpecRepository) List(ctx context.Context) ([]*model.Spec, error) {
	var specs []*model.Spec
	err := r.ds.DB(ctx).Where("status != ?", "deleted").Order("category, name").Find(&specs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list specs: %w", err)
	}
	return specs, nil
}

// ListByCategory lists specs by category
func (r *SpecRepository) ListByCategory(ctx context.Context, category string) ([]*model.Spec, error) {
	var specs []*model.Spec
	err := r.ds.DB(ctx).Where("category = ? AND status != ?", category, "deleted").Order("name").Find(&specs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list specs by category: %w", err)
	}
	return specs, nil
}

// Update updates a spec
func (r *SpecRepository) Update(ctx context.Context, spec *model.Spec) error {
	return r.ds.DB(ctx).Save(spec).Error
}

// Delete soft deletes a spec by setting status to 'deleted'
func (r *SpecRepository) Delete(ctx context.Context, name string) error {
	return r.ds.DB(ctx).Model(&model.Spec{}).
		Where("name = ?", name).
		Update("status", "deleted").Error
}

// HardDelete physically deletes a spec from database
func (r *SpecRepository) HardDelete(ctx context.Context, name string) error {
	return r.ds.DB(ctx).Where("name = ?", name).Delete(&model.Spec{}).Error
}

// UpdateStatus updates spec status
func (r *SpecRepository) UpdateStatus(ctx context.Context, name string, status string) error {
	return r.ds.DB(ctx).Model(&model.Spec{}).
		Where("name = ?", name).
		Update("status", status).Error
}
