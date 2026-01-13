package constants

// Task status constants
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "PENDING"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusFailed     TaskStatus = "FAILED"
	TaskStatusCancelled  TaskStatus = "CANCELLED"
)

func (s TaskStatus) String() string {
	return string(s)
}
