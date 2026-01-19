package novita

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"waverless/pkg/config"
	"waverless/pkg/interfaces"
	"waverless/pkg/logger"
)

// clientInterface defines the interface for Novita API client (for testing)
type clientInterface interface {
	CreateEndpoint(ctx context.Context, req *CreateEndpointRequest) (*CreateEndpointResponse, error)
	GetEndpoint(ctx context.Context, endpointID string) (*GetEndpointResponse, error)
	ListEndpoints(ctx context.Context) (*ListEndpointsResponse, error)
	UpdateEndpoint(ctx context.Context, req *UpdateEndpointRequest) error
	DeleteEndpoint(ctx context.Context, endpointID string) error
}

// replicaCallbackEntry represents a registered replica callback
type replicaCallbackEntry struct {
	id       uint64
	callback interfaces.ReplicaCallback
}

// endpointState stores the last known state of an endpoint
type endpointState struct {
	DesiredReplicas   int
	ReadyReplicas     int
	AvailableReplicas int
	Status            string
}

// NovitaDeploymentProvider implements interfaces.DeploymentProvider for Novita Serverless
type NovitaDeploymentProvider struct {
	client        clientInterface
	config        *config.NovitaConfig
	specsConfig   *SpecsConfig
	endpointCache sync.Map // Cache endpoint ID mappings: name -> endpointID

	// WatchReplicas support
	replicaCallbacks     map[uint64]*replicaCallbackEntry
	replicaCallbacksLock sync.RWMutex
	nextCallbackID       uint64
	endpointStates       sync.Map // endpoint name -> *endpointState
	watcherRunning       atomic.Bool
	watcherStopCh        chan struct{}
	pollInterval         time.Duration // Configurable poll interval
}

// NewNovitaDeploymentProvider creates a new Novita deployment provider
func NewNovitaDeploymentProvider(cfg *config.Config) (interfaces.DeploymentProvider, error) {
	if !cfg.Novita.Enabled {
		return nil, fmt.Errorf("novita provider is not enabled in config")
	}

	if cfg.Novita.APIKey == "" {
		return nil, fmt.Errorf("novita API key is required")
	}

	// Initialize specs configuration
	specsConfig, err := NewSpecsConfig(cfg.Novita.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize specs config: %w", err)
	}

	client := NewClient(&cfg.Novita)

	// Set default poll interval to 10 seconds
	pollInterval := 10 * time.Second
	if cfg.Novita.PollInterval > 0 {
		pollInterval = time.Duration(cfg.Novita.PollInterval) * time.Second
	}

	return &NovitaDeploymentProvider{
		client:           client,
		config:           &cfg.Novita,
		specsConfig:      specsConfig,
		replicaCallbacks: make(map[uint64]*replicaCallbackEntry),
		watcherStopCh:    make(chan struct{}),
		pollInterval:     pollInterval,
	}, nil
}

// Deploy deploys an application to Novita serverless
func (p *NovitaDeploymentProvider) Deploy(ctx context.Context, req *interfaces.DeployRequest) (*interfaces.DeployResponse, error) {
	logger.Infof("Deploying endpoint %s to Novita", req.Endpoint)

	// Get spec from configuration
	specInfo, err := p.specsConfig.GetSpec(req.SpecName)
	if err != nil {
		return nil, fmt.Errorf("failed to get spec for %s: %w", req.SpecName, err)
	}

	// Map Waverless request to Novita request (mapper will extract platform config from spec)
	novitaReq, err := mapDeployRequestToNovita(req, specInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to map deploy request to Novita: %w", err)
	}

	// Create endpoint
	resp, err := p.client.CreateEndpoint(ctx, novitaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create Novita endpoint: %w", err)
	}

	// Cache endpoint ID mapping
	p.endpointCache.Store(req.Endpoint, resp.ID)

	logger.Infof("Successfully deployed endpoint %s to Novita (ID: %s)", req.Endpoint, resp.ID)

	return &interfaces.DeployResponse{
		Endpoint:  req.Endpoint,
		Message:   fmt.Sprintf("%s (ID: %s)", MessageDeploySuccess, resp.ID),
		CreatedAt: "", // Novita doesn't return creation time in response
	}, nil
}

// GetApp retrieves application details
func (p *NovitaDeploymentProvider) GetApp(ctx context.Context, endpoint string) (*interfaces.AppInfo, error) {
	logger.Debugf("Getting app info for endpoint %s", endpoint)

	// Get endpoint ID from cache or find it
	endpointID, err := p.getEndpointID(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	// Get endpoint details from Novita
	resp, err := p.client.GetEndpoint(ctx, endpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint from Novita: %w", err)
	}

	return mapNovitaResponseToAppInfo(resp), nil
}

// ListApps lists all applications
func (p *NovitaDeploymentProvider) ListApps(ctx context.Context) ([]*interfaces.AppInfo, error) {
	logger.Debugf("Listing all Novita endpoints")

	resp, err := p.client.ListEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints from Novita: %w", err)
	}

	apps := make([]*interfaces.AppInfo, 0, len(resp.Endpoints))
	for _, item := range resp.Endpoints {
		// Cache endpoint ID mapping
		p.endpointCache.Store(item.Name, item.ID)

		apps = append(apps, mapNovitaListItemToAppInfo(&item))
	}

	return apps, nil
}

// DeleteApp deletes application
func (p *NovitaDeploymentProvider) DeleteApp(ctx context.Context, endpoint string) error {
	logger.Infof("Deleting endpoint %s from Novita", endpoint)

	// Get endpoint ID
	endpointID, err := p.getEndpointID(ctx, endpoint)
	if err != nil {
		return err
	}

	// Delete endpoint
	if err := p.client.DeleteEndpoint(ctx, endpointID); err != nil {
		return fmt.Errorf("failed to delete endpoint from Novita: %w", err)
	}

	// Remove from cache
	p.endpointCache.Delete(endpoint)

	logger.Infof("%s %s (ID: %s)", MessageDeleteSuccess, endpoint, endpointID)
	return nil
}

// ScaleApp scales application replicas
func (p *NovitaDeploymentProvider) ScaleApp(ctx context.Context, endpoint string, replicas int) error {
	logger.Infof("Scaling endpoint %s to %d replicas", endpoint, replicas)

	// Get endpoint ID
	endpointID, err := p.getEndpointID(ctx, endpoint)
	if err != nil {
		return err
	}

	// Get current configuration
	currentConfig, err := p.client.GetEndpoint(ctx, endpointID)
	if err != nil {
		return fmt.Errorf("failed to get current endpoint config: %w", err)
	}

	// Create update request with modified replicas
	replicasPtr := &replicas
	scaleReq := &interfaces.UpdateDeploymentRequest{
		Endpoint: endpoint,
		Replicas: replicasPtr,
	}

	updateReq := mapUpdateRequestToNovita(endpointID, scaleReq, currentConfig)
	if updateReq == nil {
		return fmt.Errorf("failed to create scale request")
	}

	if err := p.client.UpdateEndpoint(ctx, updateReq); err != nil {
		return fmt.Errorf("failed to scale endpoint: %w", err)
	}

	logger.Infof("Successfully scaled endpoint %s to %d replicas", endpoint, replicas)
	return nil
}

// GetAppStatus retrieves application status
func (p *NovitaDeploymentProvider) GetAppStatus(ctx context.Context, endpoint string) (*interfaces.AppStatus, error) {
	logger.Debugf("Getting status for endpoint %s", endpoint)

	// Get endpoint ID
	endpointID, err := p.getEndpointID(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	// Get endpoint details
	resp, err := p.client.GetEndpoint(ctx, endpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint status: %w", err)
	}

	return mapNovitaStatusToAppStatus(endpoint, &resp.Endpoint), nil
}

// GetAppLogs retrieves application logs (not supported by Novita)
func (p *NovitaDeploymentProvider) GetAppLogs(ctx context.Context, endpoint string, lines int, podName ...string) (string, error) {
	return "", fmt.Errorf(MessageLogsNotSupported)
}

// UpdateDeployment updates deployment
func (p *NovitaDeploymentProvider) UpdateDeployment(ctx context.Context, req *interfaces.UpdateDeploymentRequest) (*interfaces.DeployResponse, error) {
	logger.Infof("Updating deployment for endpoint %s", req.Endpoint)

	// Get endpoint ID
	endpointID, err := p.getEndpointID(ctx, req.Endpoint)
	if err != nil {
		return nil, err
	}

	// Get current configuration
	currentConfig, err := p.client.GetEndpoint(ctx, endpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current endpoint config: %w", err)
	}

	// Map update request
	updateReq := mapUpdateRequestToNovita(endpointID, req, currentConfig)
	if updateReq == nil {
		return nil, fmt.Errorf("failed to map update request")
	}

	// Update endpoint
	if err := p.client.UpdateEndpoint(ctx, updateReq); err != nil {
		return nil, fmt.Errorf("failed to update endpoint: %w", err)
	}

	logger.Infof("Successfully updated endpoint %s", req.Endpoint)

	return &interfaces.DeployResponse{
		Endpoint:  req.Endpoint,
		Message:   MessageUpdateSuccess,
		CreatedAt: "",
	}, nil
}

// ListSpecs lists available specifications
func (p *NovitaDeploymentProvider) ListSpecs(ctx context.Context) ([]*interfaces.SpecInfo, error) {
	return p.specsConfig.ListSpecs(), nil
}

// GetSpec retrieves specification details
func (p *NovitaDeploymentProvider) GetSpec(ctx context.Context, specName string) (*interfaces.SpecInfo, error) {
	return p.specsConfig.GetSpec(specName)
}

// PreviewDeploymentYAML previews deployment configuration (returns JSON for Novita)
func (p *NovitaDeploymentProvider) PreviewDeploymentYAML(ctx context.Context, req *interfaces.DeployRequest) (string, error) {
	// Get spec from configuration
	specInfo, err := p.specsConfig.GetSpec(req.SpecName)
	if err != nil {
		return "", fmt.Errorf("failed to get spec for %s: %w", req.SpecName, err)
	}

	region := req.Labels[LabelKeyRegion]
	if region == "" {
		region = DefaultRegion
	}

	// Map to Novita request (mapper will extract platform config from spec)
	novitaReq, err := mapDeployRequestToNovita(req, specInfo)
	if err != nil {
		return "", fmt.Errorf("failed to map deploy request to Novita: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(novitaReq, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Novita config: %w", err)
	}

	return string(jsonData), nil
}

// WatchReplicas watches replica count changes using polling mechanism
func (p *NovitaDeploymentProvider) WatchReplicas(ctx context.Context, callback interfaces.ReplicaCallback) error {
	if callback == nil {
		return fmt.Errorf("replica callback is nil")
	}

	// Register callback
	p.replicaCallbacksLock.Lock()
	callbackID := atomic.AddUint64(&p.nextCallbackID, 1)
	p.replicaCallbacks[callbackID] = &replicaCallbackEntry{
		id:       callbackID,
		callback: callback,
	}
	p.replicaCallbacksLock.Unlock()

	logger.Infof("Registered replica watch callback (ID: %d) for Novita endpoints", callbackID)

	// Start watcher if not already running
	if p.watcherRunning.CompareAndSwap(false, true) {
		logger.Infof("Starting Novita replica watcher (poll interval: %v)", p.pollInterval)
		go p.runReplicaWatcher(ctx)
	}

	// Unregister callback when context is done
	go func() {
		<-ctx.Done()
		p.replicaCallbacksLock.Lock()
		delete(p.replicaCallbacks, callbackID)
		p.replicaCallbacksLock.Unlock()
		logger.Infof("Unregistered replica watch callback (ID: %d)", callbackID)
	}()

	return nil
}

// runReplicaWatcher runs the polling loop to monitor replica changes
func (p *NovitaDeploymentProvider) runReplicaWatcher(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	logger.Infof("Novita replica watcher started")

	for {
		select {
		case <-ctx.Done():
			logger.Infof("Novita replica watcher stopped (context done)")
			p.watcherRunning.Store(false)
			return
		case <-p.watcherStopCh:
			logger.Infof("Novita replica watcher stopped (stop signal)")
			p.watcherRunning.Store(false)
			return
		case <-ticker.C:
			p.pollEndpointStates(ctx)
		}
	}
}

// pollEndpointStates polls all endpoints and detects state changes
func (p *NovitaDeploymentProvider) pollEndpointStates(ctx context.Context) {
	// List all endpoints
	resp, err := p.client.ListEndpoints(ctx)
	if err != nil {
		logger.Errorf("Failed to list Novita endpoints for polling: %v", err)
		return
	}

	// Process each endpoint
	for _, item := range resp.Endpoints {
		endpointName := item.Name

		// Get status from list item (includes full worker details)
		status := p.getEndpointStateFromListItem(&item)

		// Compare with cached state
		previousStateInterface, exists := p.endpointStates.Load(endpointName)

		var hasChanged bool
		if !exists {
			// New endpoint
			hasChanged = true
		} else {
			previousState := previousStateInterface.(*endpointState)
			hasChanged = p.hasStateChanged(previousState, status)
		}

		// Update cache
		p.endpointStates.Store(endpointName, status)

		// Trigger callbacks if state changed
		if hasChanged {
			logger.Debugf("Detected state change for endpoint %s: desired=%d, ready=%d, available=%d, status=%s",
				endpointName, status.DesiredReplicas, status.ReadyReplicas, status.AvailableReplicas, status.Status)

			p.triggerReplicaCallbacks(interfaces.ReplicaEvent{
				Name:              endpointName,
				DesiredReplicas:   status.DesiredReplicas,
				ReadyReplicas:     status.ReadyReplicas,
				AvailableReplicas: status.AvailableReplicas,
				Conditions:        p.buildConditions(status),
			})
		}
	}
}

// getEndpointStateFromListItem extracts state from list item
// Note: Novita's ListEndpoints API returns full endpoint details including workers
func (p *NovitaDeploymentProvider) getEndpointStateFromListItem(item *EndpointListItem) *endpointState {
	if item == nil {
		return &endpointState{}
	}

	status := mapNovitaStatusToWaverless(item.State.State)

	// Get desired replicas from worker config
	desiredReplicas := item.WorkerConfig.MaxNum

	// Count workers by state
	runningWorkers := 0
	// healthyWorkers := 0

	for _, worker := range item.Workers {
		if worker.State.State == NovitaStatusRunning {
			runningWorkers++
		}
		// if worker.Healthy {
		// 	healthyWorkers++
		// }
	}

	// Use running workers as ready replicas and available replicas
	readyReplicas := runningWorkers
	availableReplicas := runningWorkers

	return &endpointState{
		DesiredReplicas:   desiredReplicas,
		ReadyReplicas:     readyReplicas,
		AvailableReplicas: availableReplicas,
		Status:            status,
	}
}

// hasStateChanged checks if the endpoint state has changed
func (p *NovitaDeploymentProvider) hasStateChanged(previous, current *endpointState) bool {
	return previous.DesiredReplicas != current.DesiredReplicas ||
		previous.ReadyReplicas != current.ReadyReplicas ||
		previous.AvailableReplicas != current.AvailableReplicas ||
		previous.Status != current.Status
}

// buildConditions builds condition list from endpoint state
func (p *NovitaDeploymentProvider) buildConditions(state *endpointState) []interfaces.ReplicaCondition {
	conditions := []interfaces.ReplicaCondition{}

	if state.Status == StatusRunning && state.ReadyReplicas > 0 {
		conditions = append(conditions, interfaces.ReplicaCondition{
			Type:    "Available",
			Status:  "True",
			Reason:  "MinimumReplicasAvailable",
			Message: "Endpoint has minimum availability",
		})
	} else if state.Status == StatusPending || state.Status == StatusCreating {
		conditions = append(conditions, interfaces.ReplicaCondition{
			Type:    "Progressing",
			Status:  "True",
			Reason:  "NewEndpointAvailable",
			Message: "Endpoint is being created",
		})
	} else if state.Status == StatusFailed {
		conditions = append(conditions, interfaces.ReplicaCondition{
			Type:    "Available",
			Status:  "False",
			Reason:  "EndpointFailed",
			Message: "Endpoint has failed",
		})
	}

	return conditions
}

// triggerReplicaCallbacks triggers all registered callbacks with the event
func (p *NovitaDeploymentProvider) triggerReplicaCallbacks(event interfaces.ReplicaEvent) {
	p.replicaCallbacksLock.RLock()
	defer p.replicaCallbacksLock.RUnlock()

	for _, entry := range p.replicaCallbacks {
		// Call callback in a goroutine to avoid blocking
		go func(cb interfaces.ReplicaCallback, e interfaces.ReplicaEvent) {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic in replica callback: %v", r)
				}
			}()
			cb(e)
		}(entry.callback, event)
	}
}

// StopReplicaWatcher stops the replica watcher
func (p *NovitaDeploymentProvider) StopReplicaWatcher() {
	if p.watcherRunning.Load() {
		close(p.watcherStopCh)
	}
}

// GetPods retrieves all Pod information (not supported by Novita)
func (p *NovitaDeploymentProvider) GetPods(ctx context.Context, endpoint string) ([]*interfaces.PodInfo, error) {
	return nil, fmt.Errorf(MessagePodsNotSupported)
}

// DescribePod retrieves detailed Pod information (not supported by Novita)
func (p *NovitaDeploymentProvider) DescribePod(ctx context.Context, endpoint string, podName string) (*interfaces.PodDetail, error) {
	return nil, fmt.Errorf("DescribePod %s", MessageNotSupported)
}

// GetPodYAML retrieves Pod YAML (not supported by Novita)
func (p *NovitaDeploymentProvider) GetPodYAML(ctx context.Context, endpoint string, podName string) (string, error) {
	return "", fmt.Errorf("GetPodYAML %s", MessageNotSupported)
}

// ListPVCs lists all PersistentVolumeClaims (not supported by Novita)
func (p *NovitaDeploymentProvider) ListPVCs(ctx context.Context) ([]*interfaces.PVCInfo, error) {
	return nil, fmt.Errorf("ListPVCs %s", MessageNotSupported)
}

// GetDefaultEnv retrieves default environment variables
func (p *NovitaDeploymentProvider) GetDefaultEnv(ctx context.Context) (map[string]string, error) {
	// Return default environment variables for Novita
	defaultEnv := map[string]string{
		EnvKeyNovitaProvider: EnvValueTrue,
		EnvKeyProviderType:   EnvValueNovita,
	}

	return defaultEnv, nil
}

// getEndpointID retrieves the Novita endpoint ID for a given endpoint name
// It first checks the cache, then queries the API if not found
func (p *NovitaDeploymentProvider) getEndpointID(ctx context.Context, endpoint string) (string, error) {
	// Check cache first
	if id, ok := p.endpointCache.Load(endpoint); ok {
		return id.(string), nil
	}

	// Not in cache, query API
	logger.Debugf("Endpoint ID not in cache, querying Novita API for %s", endpoint)

	resp, err := p.client.ListEndpoints(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Find matching endpoint and update cache
	for _, item := range resp.Endpoints {
		p.endpointCache.Store(item.Name, item.ID)
		if item.Name == endpoint {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("endpoint %s not found in Novita", endpoint)
}
