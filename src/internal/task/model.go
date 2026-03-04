package task

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusRunning   Status = "RUNNING"
	StatusSucceeded Status = "SUCCEEDED"
	StatusFailed    Status = "FAILED"
	StatusCanceled  Status = "CANCELED"
)

type Task struct {
	TaskID      string     `json:"task_id"`
	TaskType    string     `json:"task_type"`
	Status      Status     `json:"status"`
	RequestJSON string     `json:"request_json,omitempty"`
	ExitCode    *int       `json:"exit_code,omitempty"`
	StdoutTail  string     `json:"stdout_tail,omitempty"`
	StderrTail  string     `json:"stderr_tail,omitempty"`
	LogPath     string     `json:"log_path,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

type ListFilter struct {
	Status    Status
	TaskType  string
	CreatedBy string
	Limit     int
	Offset    int
}
