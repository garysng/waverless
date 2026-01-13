package constants

// Worker status constants
type WorkerStatus string

const (
	WorkerStatusStarting WorkerStatus = "STARTING" // Pod created, waiting for heartbeat
	WorkerStatusOnline   WorkerStatus = "ONLINE"   // Normal operation
	WorkerStatusBusy     WorkerStatus = "BUSY"     // Processing tasks
	WorkerStatusDraining WorkerStatus = "DRAINING" // Pod terminating, no new tasks
	WorkerStatusOffline  WorkerStatus = "OFFLINE"  // Disconnected
)

func (s WorkerStatus) String() string {
	return string(s)
}
