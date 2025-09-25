package models

import (
	"encoding/json"
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// TaskType represents different types of tasks that can be processed
type TaskType string

const (
	TaskTypeEmail       TaskType = "email"
	TaskTypeImageResize TaskType = "image_resize"
	TaskTypeDataProcess TaskType = "data_process"
	TaskTypeWebhook     TaskType = "webhook"
)

// Task represents a unit of work to be processed
type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Status      TaskStatus             `json:"status"`
	Priority    int                    `json:"priority"` // Higher number = higher priority
	MaxRetries  int                    `json:"max_retries"`
	RetryCount  int                    `json:"retry_count"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	WorkerID    string                 `json:"worker_id,omitempty"`
}

// TaskResult represents the result of task processing
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	WorkerID  string                 `json:"worker_id"`
}

// NewTask creates a new task with default values
func NewTask(taskType TaskType, payload map[string]interface{}) *Task {
	return &Task{
		ID:         generateTaskID(),
		Type:       taskType,
		Payload:    payload,
		Status:     TaskStatusPending,
		Priority:   0,
		MaxRetries: 3,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}
}

// ToJSON converts task to JSON string
func (t *Task) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	return string(data), err
}

// FromJSON creates task from JSON string
func TaskFromJSON(data string) (*Task, error) {
	var task Task
	err := json.Unmarshal([]byte(data), &task)
	return &task, err
}

// CanRetry checks if the task can be retried
func (t *Task) CanRetry() bool {
	return t.RetryCount < t.MaxRetries
}

// MarkStarted marks the task as started
func (t *Task) MarkStarted(workerID string) {
	now := time.Now()
	t.Status = TaskStatusProcessing
	t.StartedAt = &now
	t.WorkerID = workerID
}

// MarkCompleted marks the task as completed
func (t *Task) MarkCompleted() {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
}

// MarkFailed marks the task as failed
func (t *Task) MarkFailed(err string) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.CompletedAt = &now
	t.Error = err
	t.RetryCount++
}