package interfaces

import "time"

type CapacityStatus string

const (
	CapacityAvailable CapacityStatus = "available"
	CapacityLimited   CapacityStatus = "limited"
	CapacitySoldOut   CapacityStatus = "sold_out"
)

// CapacityEvent 容量变更事件
type CapacityEvent struct {
	SpecName  string         `json:"specName"`
	Status    CapacityStatus `json:"status"`
	Reason    string         `json:"reason,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// SpecWithCapacity 带容量状态的规格信息
type SpecWithCapacity struct {
	*SpecInfo
	Capacity     CapacityStatus `json:"capacity"`
	SpotScore    int            `json:"spotScore,omitempty"`
	SpotPrice    float64        `json:"spotPrice,omitempty"`
	RunningCount int            `json:"runningCount,omitempty"`
	PendingCount int            `json:"pendingCount,omitempty"`
}
