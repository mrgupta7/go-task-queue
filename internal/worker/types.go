package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mridul/go-task-queue/pkg/models"
)

// WorkerPoolStatus represents the current status of the worker pool
type WorkerPoolStatus struct {
	Running     bool           `json:"running"`
	WorkerCount int            `json:"worker_count"`
	QueueSize   int            `json:"queue_size"`
	Workers     []WorkerStatus `json:"workers"`
}

// WorkerStatus represents the status of an individual worker
type WorkerStatus struct {
	ID      string `json:"id"`
	Running bool   `json:"running"`
}

// DefaultTaskProcessor is a sample implementation of TaskProcessor
// It simulates different types of work based on task type
type DefaultTaskProcessor struct{}

// NewDefaultTaskProcessor creates a new default task processor
func NewDefaultTaskProcessor() *DefaultTaskProcessor {
	return &DefaultTaskProcessor{}
}

// ProcessTask processes a task based on its type
func (p *DefaultTaskProcessor) ProcessTask(ctx context.Context, task *models.Task) *models.TaskResult {
	result := &models.TaskResult{
		TaskID:  task.ID,
		Success: true,
	}

	// Simulate different processing based on task type
	switch task.Type {
	case models.TaskTypeEmail:
		result = p.processEmailTask(ctx, task)
	case models.TaskTypeImageResize:
		result = p.processImageResizeTask(ctx, task)
	case models.TaskTypeDataProcess:
		result = p.processDataTask(ctx, task)
	case models.TaskTypeWebhook:
		result = p.processWebhookTask(ctx, task)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("unknown task type: %s", task.Type)
	}

	return result
}

// processEmailTask simulates email sending
func (p *DefaultTaskProcessor) processEmailTask(ctx context.Context, task *models.Task) *models.TaskResult {
	result := &models.TaskResult{TaskID: task.ID, Success: true}
	
	// Extract email details from payload
	recipient, ok := task.Payload["recipient"].(string)
	if !ok {
		result.Success = false
		result.Error = "missing recipient in email task"
		return result
	}
	
	subject, _ := task.Payload["subject"].(string)
	
	log.Printf("Sending email to %s with subject: %s", recipient, subject)
	
	// Simulate email sending delay (1-3 seconds)
	processingTime := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
	
	select {
	case <-time.After(processingTime):
		// Simulate occasional failures (5% chance)
		if rand.Float32() < 0.05 {
			result.Success = false
			result.Error = "SMTP server connection failed"
		} else {
			result.Result = map[string]interface{}{
				"message_id": fmt.Sprintf("msg_%d", time.Now().Unix()),
				"status":     "sent",
			}
			log.Printf("Email sent successfully to %s", recipient)
		}
	case <-ctx.Done():
		result.Success = false
		result.Error = "email task cancelled"
	}
	
	return result
}

// processImageResizeTask simulates image processing
func (p *DefaultTaskProcessor) processImageResizeTask(ctx context.Context, task *models.Task) *models.TaskResult {
	result := &models.TaskResult{TaskID: task.ID, Success: true}
	
	imageURL, ok := task.Payload["image_url"].(string)
	if !ok {
		result.Success = false
		result.Error = "missing image_url in resize task"
		return result
	}
	
	width, _ := task.Payload["width"].(float64)
	height, _ := task.Payload["height"].(float64)
	
	log.Printf("Resizing image %s to %vx%v", imageURL, width, height)
	
	// Simulate image processing delay (2-5 seconds)
	processingTime := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
	
	select {
	case <-time.After(processingTime):
		// Simulate occasional failures (10% chance)
		if rand.Float32() < 0.10 {
			result.Success = false
			result.Error = "image processing failed: unsupported format"
		} else {
			result.Result = map[string]interface{}{
				"output_url": fmt.Sprintf("%s_resized_%vx%v", imageURL, int(width), int(height)),
				"file_size":  rand.Intn(1000000) + 50000, // Random file size
			}
			log.Printf("Image resized successfully: %s", imageURL)
		}
	case <-ctx.Done():
		result.Success = false
		result.Error = "image resize task cancelled"
	}
	
	return result
}

// processDataTask simulates data processing
func (p *DefaultTaskProcessor) processDataTask(ctx context.Context, task *models.Task) *models.TaskResult {
	result := &models.TaskResult{TaskID: task.ID, Success: true}
	
	dataSource, ok := task.Payload["data_source"].(string)
	if !ok {
		result.Success = false
		result.Error = "missing data_source in data processing task"
		return result
	}
	
	operation, _ := task.Payload["operation"].(string)
	
	log.Printf("Processing data from %s with operation: %s", dataSource, operation)
	
	// Simulate data processing delay (3-8 seconds)
	processingTime := time.Duration(3000+rand.Intn(5000)) * time.Millisecond
	
	select {
	case <-time.After(processingTime):
		// Simulate occasional failures (8% chance)
		if rand.Float32() < 0.08 {
			result.Success = false
			result.Error = "data processing failed: invalid data format"
		} else {
			result.Result = map[string]interface{}{
				"records_processed": rand.Intn(10000) + 1000,
				"output_file":       fmt.Sprintf("/tmp/processed_%d.csv", time.Now().Unix()),
				"processing_time":   processingTime.String(),
			}
			log.Printf("Data processing completed for: %s", dataSource)
		}
	case <-ctx.Done():
		result.Success = false
		result.Error = "data processing task cancelled"
	}
	
	return result
}

// processWebhookTask simulates webhook calls
func (p *DefaultTaskProcessor) processWebhookTask(ctx context.Context, task *models.Task) *models.TaskResult {
	result := &models.TaskResult{TaskID: task.ID, Success: true}
	
	webhookURL, ok := task.Payload["url"].(string)
	if !ok {
		result.Success = false
		result.Error = "missing url in webhook task"
		return result
	}
	
	method, _ := task.Payload["method"].(string)
	if method == "" {
		method = "POST"
	}
	
	log.Printf("Calling webhook %s with method %s", webhookURL, method)
	
	// Simulate webhook call delay (1-4 seconds)
	processingTime := time.Duration(1000+rand.Intn(3000)) * time.Millisecond
	
	select {
	case <-time.After(processingTime):
		// Simulate occasional failures (12% chance)
		if rand.Float32() < 0.12 {
			result.Success = false
			result.Error = "webhook call failed: connection timeout"
		} else {
			statusCode := 200
			if rand.Float32() < 0.05 {
				statusCode = 500 // 5% chance of 500 error
			}
			
			result.Result = map[string]interface{}{
				"status_code":    statusCode,
				"response_time":  processingTime.Milliseconds(),
				"response_body":  `{"status": "success", "message": "webhook processed"}`,
			}
			log.Printf("Webhook call completed: %s (status: %d)", webhookURL, statusCode)
		}
	case <-ctx.Done():
		result.Success = false
		result.Error = "webhook task cancelled"
	}
	
	return result
}