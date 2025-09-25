package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mridul/go-task-queue/internal/queue"
	"github.com/mridul/go-task-queue/internal/worker"
	"github.com/mridul/go-task-queue/pkg/models"
)

// Server represents the HTTP API server
type Server struct {
	queue      queue.Queue
	workerPool *worker.WorkerPool
	router     *mux.Router
	httpServer *http.Server
	port       int
}

// TaskRequest represents a request to create a new task
type TaskRequest struct {
	Type       models.TaskType        `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Priority   int                    `json:"priority"`
	MaxRetries int                    `json:"max_retries"`
}

// TaskResponse represents a task response
type TaskResponse struct {
	*models.Task
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

// StatsResponse represents system statistics
type StatsResponse struct {
	QueueSize     int                     `json:"queue_size"`
	WorkerPool    worker.WorkerPoolStatus `json:"worker_pool"`
	TasksTotal    int                     `json:"tasks_total"`
	TasksPending  int                     `json:"tasks_pending"`
	TasksRunning  int                     `json:"tasks_running"`
	TasksComplete int                     `json:"tasks_complete"`
	TasksFailed   int                     `json:"tasks_failed"`
}

// NewServer creates a new API server
func NewServer(q queue.Queue, wp *worker.WorkerPool, port int) *Server {
	s := &Server{
		queue:      q,
		workerPool: wp,
		port:       port,
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Task management
	api.HandleFunc("/tasks", s.createTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", s.getTask).Methods("GET")
	api.HandleFunc("/tasks", s.listTasks).Methods("GET")

	// System endpoints
	api.HandleFunc("/health", s.healthCheck).Methods("GET")
	api.HandleFunc("/stats", s.getStats).Methods("GET")
	api.HandleFunc("/workers", s.getWorkerStatus).Methods("GET")

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.jsonMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP server on port %d", s.port)

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Printf("Shutting down HTTP server...")

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// createTask handles POST /api/v1/tasks
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Validate task type
	if req.Type == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Missing task type", "task type is required")
		return
	}

	// Create task
	task := models.NewTask(req.Type, req.Payload)
	if req.Priority > 0 {
		task.Priority = req.Priority
	}
	if req.MaxRetries > 0 {
		task.MaxRetries = req.MaxRetries
	}

	// Enqueue task
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.queue.Enqueue(ctx, task); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to enqueue task", err.Error())
		return
	}

	log.Printf("Task %s created successfully (type: %s)", task.ID, task.Type)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TaskResponse{Task: task})
}

// getTask handles GET /api/v1/tasks/{id}
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := s.queue.GetTask(ctx, taskID)
	if err != nil {
		if err == queue.ErrTaskNotFound {
			s.writeErrorResponse(w, http.StatusNotFound, "Task not found", fmt.Sprintf("task %s not found", taskID))
		} else {
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get task", err.Error())
		}
		return
	}

	json.NewEncoder(w).Encode(TaskResponse{Task: task})
}

// listTasks handles GET /api/v1/tasks
func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	// This is a simplified implementation
	// In a real system, you'd want pagination, filtering, etc.

	response := map[string]interface{}{
		"message":    "Task listing not implemented in memory queue",
		"queue_size": s.queue.Size(),
		"note":       "Use Redis queue for full task listing functionality",
	}

	json.NewEncoder(w).Encode(response)
}

// healthCheck handles GET /api/v1/health
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    "N/A",
	}

	json.NewEncoder(w).Encode(response)
}

// getStats handles GET /api/v1/stats
func (s *Server) getStats(w http.ResponseWriter, r *http.Request) {
	workerStatus := s.workerPool.GetStatus()

	response := StatsResponse{
		QueueSize:     s.queue.Size(),
		WorkerPool:    workerStatus,
		TasksTotal:    0, // Would be tracked in production
		TasksPending:  s.queue.Size(),
		TasksRunning:  0, // Would be tracked in production
		TasksComplete: 0, // Would be tracked in production
		TasksFailed:   0, // Would be tracked in production
	}

	json.NewEncoder(w).Encode(response)
}

// getWorkerStatus handles GET /api/v1/workers
func (s *Server) getWorkerStatus(w http.ResponseWriter, r *http.Request) {
	status := s.workerPool.GetStatus()
	json.NewEncoder(w).Encode(status)
}

// writeErrorResponse writes a standardized error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.WriteHeader(statusCode)
	response := ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Message: details,
	}
	json.NewEncoder(w).Encode(response)
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start))
	})
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonMiddleware sets JSON content type
func (s *Server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
