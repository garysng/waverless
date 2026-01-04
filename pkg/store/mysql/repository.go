package mysql

import "waverless/pkg/config"

// Repository aggregates all MySQL repositories
type Repository struct {
	ds *Datastore

	Endpoint         *EndpointRepository
	Task             *TaskRepository
	TaskEvent        *TaskEventRepository
	TaskStatistics   *TaskStatisticsRepository
	ScalingEvent     *ScalingEventRepository
	AutoscalerConfig *AutoscalerConfigRepository
	GPUUsage         *GPUUsageRepository
	Spec             *SpecRepository
}

// NewRepository creates a new MySQL repository with all sub-repositories
func NewRepository(dsn string, proxyConfig *config.ProxyConfig) (*Repository, error) {
	ds, err := NewDatastore(dsn, proxyConfig)
	if err != nil {
		return nil, err
	}

	return &Repository{
		ds:               ds,
		Endpoint:         NewEndpointRepository(ds),
		Task:             NewTaskRepository(ds),
		TaskEvent:        NewTaskEventRepository(ds),
		TaskStatistics:   NewTaskStatisticsRepository(ds),
		ScalingEvent:     NewScalingEventRepository(ds),
		AutoscalerConfig: NewAutoscalerConfigRepository(ds),
		GPUUsage:         NewGPUUsageRepository(ds),
		Spec:             NewSpecRepository(ds),
	}, nil
}

// GetDatastore returns the underlying datastore for transaction support
func (r *Repository) GetDatastore() *Datastore {
	return r.ds
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.ds.Close()
}
