package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mridul/go-task-queue/internal/queue"
	"github.com/mridul/go-task-queue/pkg/models"
)

// TaskProcessor defines the interface for processing tasks
type TaskProcessor interface {
	ProcessTask(ctx context.Context, task *models.Task) *models.TaskResult
}

// WorkerPool manages a pool of workers that process tasks concurrently
type WorkerPool struct {
	queue       queue.Queue
	processor   TaskProcessor
	workerCount int
	workers     []*Worker
	resultsCh   chan *models.TaskResult
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	running     bool
}

// Worker represents a single worker that processes tasks
type Worker struct {
	id          string
	pool        *WorkerPool
	queue       queue.Queue
	processor   TaskProcessor
	resultsCh   chan *models.TaskResult
	shutdownCh  chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(q queue.Queue, processor TaskProcessor, workerCount int) *WorkerPool {
	return &WorkerPool{
		queue:       q,
		processor:   processor,
		workerCount: workerCount,
		workers:     make([]*Worker, 0, workerCount),
		resultsCh:   make(chan *models.TaskResult, workerCount*2),
		shutdownCh:  make(chan struct{}),
	}
}

// Start initializes and starts all workers in the pool
func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return fmt.Errorf("worker pool is already running")
	}

	log.Printf("Starting worker pool with %d workers", wp.workerCount)

	// Create and start workers
	for i := 0; i < wp.workerCount; i++ {
		worker := wp.createWorker(i)
		wp.workers = append(wp.workers, worker)
		
		wp.wg.Add(1)
		go worker.start(ctx)
	}

	// Start result processor
	wp.wg.Add(1)
	go wp.processResults(ctx)

	wp.running = true
	log.Printf("Worker pool started successfully")
	
	return nil
}

// createWorker creates a new worker instance
func (wp *WorkerPool) createWorker(index int) *Worker {
	workerCtx, cancel := context.WithCancel(context.Background())
	
	return &Worker{
		id:         fmt.Sprintf("worker-%d-%s", index, models.GenerateWorkerID()),
		pool:       wp,
		queue:      wp.queue,
		processor:  wp.processor,
		resultsCh:  wp.resultsCh,
		shutdownCh: wp.shutdownCh,
		ctx:        workerCtx,
		cancel:     cancel,
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop(ctx context.Context) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		return nil
	}

	log.Printf("Shutting down worker pool...")

	// Signal all workers to stop
	for _, worker := range wp.workers {
		worker.cancel()
	}

	// Close shutdown channel to signal result processor
	close(wp.shutdownCh)

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Worker pool shut down gracefully")
	case <-ctx.Done():
		log.Printf("Worker pool shutdown timed out")
		return ctx.Err()
	}

	wp.running = false
	close(wp.resultsCh)
	
	return nil
}

// GetResults returns the channel for receiving task results
func (wp *WorkerPool) GetResults() <-chan *models.TaskResult {
	return wp.resultsCh
}

// GetStatus returns the current status of the worker pool
func (wp *WorkerPool) GetStatus() WorkerPoolStatus {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	status := WorkerPoolStatus{
		Running:     wp.running,
		WorkerCount: wp.workerCount,
		QueueSize:   wp.queue.Size(),
		Workers:     make([]WorkerStatus, len(wp.workers)),
	}

	for i, worker := range wp.workers {
		status.Workers[i] = WorkerStatus{
			ID:      worker.id,
			Running: wp.running, // Simplified - could track individual worker status
		}
	}

	return status
}

// processResults handles task results from workers
func (wp *WorkerPool) processResults(ctx context.Context) {
	defer wp.wg.Done()
	
	for {
		select {
		case result := <-wp.resultsCh:
			if result != nil {
				wp.handleTaskResult(ctx, result)
			}
		case <-wp.shutdownCh:
			log.Printf("Result processor shutting down")
			return
		case <-ctx.Done():
			log.Printf("Result processor cancelled")
			return
		}
	}
}

// handleTaskResult processes a completed task result
func (wp *WorkerPool) handleTaskResult(ctx context.Context, result *models.TaskResult) {
	// Update task status in queue
	task, err := wp.queue.GetTask(ctx, result.TaskID)
	if err != nil {
		log.Printf("Failed to get task %s: %v", result.TaskID, err)
		return
	}

	if result.Success {
		task.MarkCompleted()
		log.Printf("Task %s completed successfully by worker %s in %v", 
			result.TaskID, result.WorkerID, result.Duration)
	} else {
		task.MarkFailed(result.Error)
		log.Printf("Task %s failed on worker %s: %s", 
			result.TaskID, result.WorkerID, result.Error)
		
		// Retry logic
		if task.CanRetry() {
			task.Status = models.TaskStatusPending
			if err := wp.queue.Enqueue(ctx, task); err != nil {
				log.Printf("Failed to re-enqueue task %s for retry: %v", task.ID, err)
			} else {
				log.Printf("Task %s re-enqueued for retry (%d/%d)", 
					task.ID, task.RetryCount, task.MaxRetries)
			}
		}
	}

	// Update task in queue
	if err := wp.queue.UpdateTask(ctx, task); err != nil {
		log.Printf("Failed to update task %s: %v", task.ID, err)
	}
}

// start begins the worker's task processing loop
func (w *Worker) start(ctx context.Context) {
	defer w.pool.wg.Done()
	
	log.Printf("Worker %s started", w.id)
	
	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Worker %s shutting down", w.id)
			return
		case <-ctx.Done():
			log.Printf("Worker %s cancelled", w.id)
			return
		default:
			w.processNextTask(ctx)
		}
	}
}

// processNextTask gets and processes the next available task
func (w *Worker) processNextTask(ctx context.Context) {
	// Add timeout for dequeue operation
	dequeueCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	task, err := w.queue.Dequeue(dequeueCtx)
	if err != nil {
		if err != context.DeadlineExceeded {
			log.Printf("Worker %s failed to dequeue task: %v", w.id, err)
		}
		return
	}

	if task == nil {
		// No task available, sleep briefly to avoid busy waiting
		time.Sleep(100 * time.Millisecond)
		return
	}

	// Mark task as started
	task.MarkStarted(w.id)
	if err := w.queue.UpdateTask(ctx, task); err != nil {
		log.Printf("Worker %s failed to update task %s: %v", w.id, task.ID, err)
		return
	}

	log.Printf("Worker %s processing task %s (type: %s)", w.id, task.ID, task.Type)
	
	// Process the task with timeout
	startTime := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30 second task timeout
	defer cancel()
	
	result := w.processor.ProcessTask(taskCtx, task)
	result.Duration = time.Since(startTime)
	result.WorkerID = w.id

	// Send result to results channel
	select {
	case w.resultsCh <- result:
	case <-ctx.Done():
		log.Printf("Worker %s cancelled while sending result", w.id)
		return
	}
}