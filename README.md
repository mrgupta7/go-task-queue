# Go Task Queue

A distributed task queue system built in Go

A high-performance, scalable task queue system built in Go featuring concurrent processing, graceful shutdown, and multiple deployment options. This project demonstrates production-ready Go development practices including channels, context management, Docker containerization, and Kubernetes deployment.

## 🚀 Features

- **Concurrent Processing**: Worker pools with configurable concurrency using Go channels and goroutines
- **In-Memory Queue**: High-performance task queuing with priority support
- **Redis Integration**: 🚧 *In Progress* - Distributed queuing for multi-instance deployment
- **Priority Queuing**: Support for high-priority task processing
- **Graceful Shutdown**: Context-based cancellation and clean shutdown handling
- **REST API**: HTTP endpoints for task submission and status monitoring
- **Health Checks**: Built-in health monitoring and metrics endpoints
- **Retry Logic**: Configurable task retry with exponential backoff
- **Docker Support**: Multi-stage builds and production-ready containers
- **Kubernetes Ready**: Complete K8s manifests with auto-scaling and load balancing
- **Production Features**: Rate limiting, CORS, logging middleware

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│   API Server     │───▶│   Task Queue    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │                        │
                               ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Monitoring    │◀───│   Worker Pool    │◀───│  Redis/Memory   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │
                               ▼
                    ┌──────────────────┐
                    │ Task Processors  │
                    └──────────────────┘
```

## 📋 Task Types Supported

- **Email**: Send emails with SMTP simulation
- **Image Processing**: Resize images with configurable dimensions
- **Data Processing**: Process CSV files and data transformations
- **Webhooks**: Make HTTP calls to external APIs
- **Custom**: Extensible processor interface for new task types

## 🛠️ Installation & Setup

### Prerequisites

- Go 1.23+
- Docker & Docker Compose (for containerized deployment)
- Kubernetes cluster (for K8s deployment)
- Redis (coming soon for distributed queues)

### Local Development

1. **Clone the repository**:
```bash
git clone <repository-url>
cd go-task-queue
```

2. **Install dependencies**:
```bash
go mod tidy
```

3. **Build the application**:
```bash
go build -o bin/task-queue cmd/server/main.go
```

4. **Run the application**:
```bash
./bin/task-queue
```

5. **Redis Integration** (Coming Soon!):
```bash
# Redis-backed distributed queuing is under development
# Currently uses high-performance in-memory queues
REDIS_URL=redis://localhost:6379 ./bin/task-queue  # Will be supported soon!
```

## 🔧 Configuration

The application is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `WORKER_COUNT` | 5 | Number of worker goroutines |
| `QUEUE_SIZE` | 1000 | In-memory queue buffer size |
| `REDIS_URL` | "" | Redis connection URL (🚧 coming soon!) |

## 📡 API Endpoints

### Create Task
```bash
POST /api/v1/tasks
Content-Type: application/json

{
  "type": "email",
  "payload": {
    "recipient": "user@example.com",
    "subject": "Hello!",
    "body": "Task queue is working!"
  },
  "priority": 0,
  "max_retries": 3
}
```

### Get Task Status
```bash
GET /api/v1/tasks/{task_id}
```

### Health Check
```bash
GET /api/v1/health
```

### System Statistics
```bash
GET /api/v1/stats
```

### Worker Status
```bash
GET /api/v1/workers
```

## 🚀 Quick Start Examples

### 1. Basic Task Creation

```bash
# Start the server
go run cmd/server/main.go

# Create an email task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": {
      "recipient": "test@example.com",
      "subject": "Welcome!",
      "body": "Hello from task queue!"
    }
  }'
```

### 2. High Priority Task

```bash
# Create a high-priority image resize task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "image_resize",
    "payload": {
      "image_url": "https://example.com/image.jpg",
      "width": 800,
      "height": 600
    },
    "priority": 5
  }'
```

### 3. Run Demo Script

```bash
# Make sure the server is running, then:
chmod +x examples/demo.sh
./examples/demo.sh
```

## 🐳 Docker Deployment

### Single Instance

```bash
# Build image
docker build -f deployments/docker/Dockerfile -t task-queue .

# Run container
docker run -p 8080:8080 \
  -e WORKER_COUNT=10 \
  -e QUEUE_SIZE=2000 \
  task-queue
```

### Docker Compose (Multi-Instance)

```bash
cd deployments/docker
docker-compose up -d
```

This starts:
- Two task queue instances (in-memory queues)
- Nginx load balancer
- 🚧 Redis integration coming soon for distributed queuing

Access points:
- **Load Balancer**: http://localhost
- **Direct Instance 1**: http://localhost:8080  
- **Direct Instance 2**: http://localhost:8081

## ☸️ Kubernetes Deployment

### Quick Deploy

```bash
# Apply all manifests
kubectl apply -f deployments/k8s/

# Check status
kubectl get pods -n task-queue-system
```

### Step-by-Step Deploy

```bash
# 1. Create namespace and configs
kubectl apply -f deployments/k8s/01-namespace-configmap.yaml

# 2. Deploy Redis
kubectl apply -f deployments/k8s/02-redis.yaml

# 3. Deploy application
kubectl apply -f deployments/k8s/03-task-queue.yaml

# 4. Setup ingress
kubectl apply -f deployments/k8s/04-ingress.yaml
```

### Scaling

```bash
# Manual scaling
kubectl scale deployment task-queue --replicas=5 -n task-queue-system

# The HPA will automatically scale between 2-10 replicas based on CPU/memory
kubectl get hpa -n task-queue-system
```

## 🔍 Monitoring & Observability

### Health Checks
- **HTTP**: `GET /api/v1/health`
- **Kubernetes**: Built-in liveness and readiness probes
- **Docker**: HEALTHCHECK instruction included

### Metrics
- Queue size and worker status via `/api/v1/stats`
- Task completion rates and error rates
- System resource utilization

### Logs
```bash
# Local development
tail -f logs/task-queue.log

# Docker
docker logs -f task-queue-container

# Kubernetes
kubectl logs -f deployment/task-queue -n task-queue-system
```

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Load Testing
```bash
# Install hey (HTTP load testing tool)
go install github.com/rakyll/hey@latest

# Load test task creation
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"type":"email","payload":{"recipient":"test@example.com","subject":"Load test"}}' \
  http://localhost:8080/api/v1/tasks
```

### Integration Testing
```bash
# Run demo script for end-to-end testing
./examples/demo.sh
```

## 🛡️ Production Considerations

### Security
- Non-root container execution
- Read-only root filesystem
- Network policies for pod-to-pod communication
- Secret management for Redis credentials

### Performance
- Horizontal pod autoscaling based on CPU/memory
- Redis clustering for high availability
- Connection pooling and circuit breakers
- Optimized Docker multi-stage builds

### Reliability
- Graceful shutdown with 30s timeout
- Task retry logic with configurable limits
- Health checks and automatic restarts
- Persistent volumes for Redis data

## 🔧 Development

### Project Structure
```
go-task-queue/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/            # HTTP API handlers
│   ├── queue/          # Queue implementations
│   └── worker/         # Worker pool and task processors
├── pkg/models/         # Shared data models
├── deployments/
│   ├── docker/         # Docker and compose files
│   └── k8s/            # Kubernetes manifests
├── examples/           # Demo scripts and examples
└── README.md
```

### Adding New Task Types

1. **Define task type** in `pkg/models/task.go`:
```go
const TaskTypeCustom TaskType = "custom_task"
```

2. **Implement processor** in `internal/worker/types.go`:
```go
func (p *DefaultTaskProcessor) processCustomTask(ctx context.Context, task *models.Task) *models.TaskResult {
    // Your implementation here
}
```

3. **Add to switch statement** in `ProcessTask` method

### Adding New Queue Backends

Implement the `Queue` interface in `internal/queue/queue.go`:
```go
type Queue interface {
    Enqueue(ctx context.Context, task *models.Task) error
    Dequeue(ctx context.Context) (*models.Task, error)
    Size() int
    Close() error
    GetTask(ctx context.Context, taskID string) (*models.Task, error)
    UpdateTask(ctx context.Context, task *models.Task) error
}
```