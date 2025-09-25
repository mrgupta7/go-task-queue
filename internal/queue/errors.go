package queue

import "errors"

var (
	// ErrQueueClosed is returned when trying to operate on a closed queue
	ErrQueueClosed = errors.New("queue is closed")
	
	// ErrTaskNotFound is returned when a task is not found in the queue
	ErrTaskNotFound = errors.New("task not found")
	
	// ErrQueueFull is returned when the queue is at capacity
	ErrQueueFull = errors.New("queue is full")
	
	// ErrInvalidTask is returned when an invalid task is provided
	ErrInvalidTask = errors.New("invalid task")
)