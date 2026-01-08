package constants

// K8s label keys
const (
	LabelApp       = "app"        // Endpoint name
	LabelManagedBy = "managed-by" // Manager identifier
	LabelComponent = "component"  // Component type

	ManagedByWaverless = "waverless"
)

// Pod phase constants (from K8s)
const (
	PodPhaseRunning   = "Running"
	PodPhasePending   = "Pending"
	PodPhaseSucceeded = "Succeeded"
	PodPhaseFailed    = "Failed"
	PodPhaseUnknown   = "Unknown"
)

// Pod status constants (custom)
const (
	PodStatusRunning     = "Running"
	PodStatusPending     = "Pending"
	PodStatusStarting    = "Starting"
	PodStatusTerminating = "Terminating"
	PodStatusFailed      = "Failed"
)
