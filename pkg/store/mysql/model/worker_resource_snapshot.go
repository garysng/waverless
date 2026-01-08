package model

import "time"

// WorkerResourceSnapshot represents a point-in-time resource usage snapshot
type WorkerResourceSnapshot struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	WorkerID        string    `gorm:"size:255;not null;index:idx_worker_snapshot,priority:1"`
	SnapshotAt      time.Time `gorm:"not null;index:idx_worker_snapshot,priority:2;index:idx_snapshot_time"`
	GPUUtilization  float64   `gorm:"type:decimal(5,2)"`
	GPUMemoryUsedMB int64
	GPUMemoryTotalMB int64
	CPUUtilization  float64   `gorm:"type:decimal(5,2)"`
	MemoryUsedMB    int64
	MemoryTotalMB   int64
	CurrentTaskID   string    `gorm:"size:255"`
	IsIdle          bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

func (WorkerResourceSnapshot) TableName() string { return "worker_resource_snapshots" }
