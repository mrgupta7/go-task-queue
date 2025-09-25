package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mridul/go-task-queue/internal/api"
	"github.com/mridul/go-task-queue/internal/queue"
	"github.com/mridul/go-task-queue/internal/worker"
)

const (
	DefaultPort        = 8080
	DefaultWorkerCount = 5
	DefaultQueueSize   = 1000
)

func main() {
	log.Println("Starting Distributed Task Queue System...")

	// Configuration from environment variables
	port := getEnvInt("PORT", DefaultPort)
	workerCount := getEnvInt("WORKER_COUNT", DefaultWorkerCount)
	queueSize := getEnvInt("QUEUE_SIZE", DefaultQueueSize)
	redisURL := getEnvString("REDIS_URL", "")

	log.Printf("Configuration: port=%d, workers=%d, queue_size=%d", port, workerCount, queueSize)

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize queue (Redis if URL provided, otherwise memory)
	var taskQueue queue.Queue
	if redisURL != "" {
		redisQueue, err := queue.NewRedisQueue(redisURL)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v. Falling back to memory queue.", err)
			taskQueue = queue.NewMemoryQueue(queueSize)
			log.Printf("Memory queue initialized with buffer size %d", queueSize)
		} else {
			taskQueue = redisQueue
			log.Printf("Redis queue initialized with URL: %s", redisURL)
		}
	} else {
		taskQueue = queue.NewMemoryQueue(queueSize)
		log.Printf("Memory queue initialized with buffer size %d", queueSize)
	}

	// Initialize task processor
	processor := worker.NewDefaultTaskProcessor()
	log.Printf("Default task processor initialized")

	// Initialize worker pool
	workerPool := worker.NewWorkerPool(taskQueue, processor, workerCount)
	log.Printf("Worker pool initialized with %d workers", workerCount)

	// Initialize HTTP server
	server := api.NewServer(taskQueue, workerPool, port)
	log.Printf("HTTP server initialized on port %d", port)

	// Start worker pool
	if err := workerPool.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker pool: %v", err)
	}

	// Start HTTP server
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Monitor results (optional - for demonstration)
	go monitorResults(ctx, workerPool)

	log.Println("✅ Distributed Task Queue System started successfully!")
	log.Println("📡 API endpoints available:")
	log.Printf("   - POST   http://localhost:%d/api/v1/tasks", port)
	log.Printf("   - GET    http://localhost:%d/api/v1/tasks/{id}", port)
	log.Printf("   - GET    http://localhost:%d/api/v1/health", port)
	log.Printf("   - GET    http://localhost:%d/api/v1/stats", port)
	log.Printf("   - GET    http://localhost:%d/api/v1/workers", port)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🔄 Shutdown signal received, initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	log.Println("Stopping HTTP server...")
	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}

	// Stop worker pool
	log.Println("Stopping worker pool...")
	if err := workerPool.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping worker pool: %v", err)
	}

	// Close queue
	log.Println("Closing task queue...")
	if err := taskQueue.Close(); err != nil {
		log.Printf("Error closing queue: %v", err)
	}

	log.Println("✅ Distributed Task Queue System shut down gracefully")
}

// monitorResults monitors task results for demonstration purposes
func monitorResults(ctx context.Context, workerPool *worker.WorkerPool) {
	results := workerPool.GetResults()

	for {
		select {
		case result := <-results:
			if result == nil {
				return // Channel closed
			}

			status := "✅ COMPLETED"
			if !result.Success {
				status = "❌ FAILED"
			}

			log.Printf("📊 Task Result: %s - Task %s (%t) processed by %s in %v",
				status, result.TaskID, result.Success, result.WorkerID, result.Duration)

		case <-ctx.Done():
			log.Println("Result monitor shutting down")
			return
		}
	}
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvString gets a string environment variable with a default value
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
