package queue

import (
	"context"
	"sync"

	"github.com/mridul/go-task-queue/pkg/models"
)

// Queue interface defines the contract for task queue implementations
type Queue interface {
	// Enqueue adds a task to the queue
	Enqueue(ctx context.Context, task *models.Task) error
	
	// Dequeue removes and returns a task from the queue
	// Returns nil if no task is available
	Dequeue(ctx context.Context) (*models.Task, error)
	
	// Size returns the current number of tasks in the queue
	Size() int
	
	// Close gracefully shuts down the queue
	Close() error
	
	// GetTask retrieves a task by ID without removing it
	GetTask(ctx context.Context, taskID string) (*models.Task, error)
	
	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *models.Task) error
}

// MemoryQueue implements Queue interface using Go channels and maps
type MemoryQueue struct {
	tasks       chan *models.Task           // Channel for task queue
	taskStore   map[string]*models.Task     // Store for task lookup by ID
	priorities  map[int][]*models.Task      // Priority-based task storage
	mu          sync.RWMutex                // Mutex for thread-safe operations
	closed      bool                        // Flag to track if queue is closed
	closeCh     chan struct{}               // Channel to signal closure
}

// NewMemoryQueue creates a new in-memory queue with the specified buffer size
func NewMemoryQueue(bufferSize int) *MemoryQueue {
	return &MemoryQueue{
		tasks:      make(chan *models.Task, bufferSize),
		taskStore:  make(map[string]*models.Task),
		priorities: make(map[int][]*models.Task),
		closeCh:    make(chan struct{}),
	}
}

// Enqueue adds a task to the queue
func (mq *MemoryQueue) Enqueue(ctx context.Context, task *models.Task) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if mq.closed {
		return ErrQueueClosed
	}
	
	// Store task for lookup
	mq.taskStore[task.ID] = task
	
	// Handle priority queuing
	if task.Priority > 0 {
		mq.priorities[task.Priority] = append(mq.priorities[task.Priority], task)
		// Try to send high priority task immediately
		select {
		case mq.tasks <- task:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// If channel is full, task stays in priority queue
			return nil
		}
	}
	
	// For normal priority tasks, send to channel
	select {
	case mq.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-mq.closeCh:
		return ErrQueueClosed
	}
}

// Dequeue removes and returns a task from the queue
func (mq *MemoryQueue) Dequeue(ctx context.Context) (*models.Task, error) {
	// First check for high priority tasks
	if task := mq.getHighPriorityTask(); task != nil {
		return task, nil
	}
	
	// Then check regular queue
	select {
	case task := <-mq.tasks:
		if task != nil {
			return task, nil
		}
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mq.closeCh:
		return nil, ErrQueueClosed
	default:
		return nil, nil // No task available
	}
}

// getHighPriorityTask retrieves the highest priority task available
func (mq *MemoryQueue) getHighPriorityTask() *models.Task {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	// Find highest priority with available tasks
	highestPriority := -1
	for priority := range mq.priorities {
		if priority > highestPriority && len(mq.priorities[priority]) > 0 {
			highestPriority = priority
		}
	}
	
	if highestPriority == -1 {
		return nil
	}
	
	// Get and remove the first task from highest priority
	tasks := mq.priorities[highestPriority]
	if len(tasks) == 0 {
		return nil
	}
	
	task := tasks[0]
	mq.priorities[highestPriority] = tasks[1:]
	
	return task
}

// Size returns the current number of tasks in the queue
func (mq *MemoryQueue) Size() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	size := len(mq.tasks)
	
	// Add priority queue sizes
	for _, tasks := range mq.priorities {
		size += len(tasks)
	}
	
	return size
}

// GetTask retrieves a task by ID without removing it
func (mq *MemoryQueue) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	if task, exists := mq.taskStore[taskID]; exists {
		return task, nil
	}
	
	return nil, ErrTaskNotFound
}

// UpdateTask updates an existing task
func (mq *MemoryQueue) UpdateTask(ctx context.Context, task *models.Task) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if mq.closed {
		return ErrQueueClosed
	}
	
	if _, exists := mq.taskStore[task.ID]; !exists {
		return ErrTaskNotFound
	}
	
	mq.taskStore[task.ID] = task
	return nil
}

// Close gracefully shuts down the queue
func (mq *MemoryQueue) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if !mq.closed {
		mq.closed = true
		close(mq.closeCh)
		close(mq.tasks)
	}
	
	return nil
}