package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// generateTaskID generates a unique task ID
func generateTaskID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("task_%d_%s", timestamp, hex.EncodeToString(randomBytes))
}

// GenerateWorkerID generates a unique worker ID
func GenerateWorkerID() string {
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("worker_%s", hex.EncodeToString(randomBytes))
}