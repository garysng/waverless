package capacity

import (
	"context"

	"waverless/pkg/interfaces"
)

// GenericProvider 通用 fallback 实现，不主动感知容量
type GenericProvider struct{}

func NewGenericProvider() *GenericProvider {
	return &GenericProvider{}
}

func (p *GenericProvider) SupportsWatch() bool { return false }

func (p *GenericProvider) Watch(ctx context.Context, callback func(interfaces.CapacityEvent)) error {
	// 不支持 watch
	<-ctx.Done()
	return nil
}

func (p *GenericProvider) Check(ctx context.Context, specName string) (*interfaces.CapacityEvent, error) {
	// 默认返回 available
	return &interfaces.CapacityEvent{
		SpecName: specName,
		Status:   interfaces.CapacityAvailable,
	}, nil
}

func (p *GenericProvider) CheckAll(ctx context.Context) ([]interfaces.CapacityEvent, error) {
	return nil, nil
}
