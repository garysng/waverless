package interfaces

import (
	"context"

	"waverless/internal/model"
)

// WorkerLister provides worker listing capability
type WorkerLister interface {
	ListWorkers(ctx context.Context, endpoint string) ([]*model.Worker, error)
	GetWorker(ctx context.Context, workerID string) (*model.Worker, error)
}
