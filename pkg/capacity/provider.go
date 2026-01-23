package capacity

import (
	"context"

	"waverless/pkg/interfaces"
)

// Provider 容量感知接口
type Provider interface {
	// SupportsWatch 是否支持 watch 模式
	SupportsWatch() bool

	// Watch 被动监听容量变化
	Watch(ctx context.Context, callback func(interfaces.CapacityEvent)) error

	// Check 主动查询某个 spec 的容量状态
	Check(ctx context.Context, specName string) (*interfaces.CapacityEvent, error)

	// CheckAll 批量查询所有 spec
	CheckAll(ctx context.Context) ([]interfaces.CapacityEvent, error)
}
