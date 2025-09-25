#!/bin/bash

# Demo script for Distributed Task Queue System
# This script demonstrates the API endpoints and task processing

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 Distributed Task Queue System Demo"
echo "======================================"

# Function to check if server is running
check_server() {
    if ! curl -s "${BASE_URL}/health" > /dev/null; then
        echo "❌ Server is not running. Please start it first with: go run cmd/server/main.go"
        exit 1
    fi
    echo "✅ Server is running"
}

# Function to create a task
create_task() {
    local task_type=$1
    local payload=$2
    local priority=$3
    
    echo "📝 Creating $task_type task..."
    
    response=$(curl -s -X POST "${BASE_URL}/tasks" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"$task_type\",
            \"payload\": $payload,
            \"priority\": $priority
        }")
    
    task_id=$(echo $response | jq -r '.id')
    echo "   Task created: $task_id"
    echo "$task_id"
}

# Function to get task status
get_task() {
    local task_id=$1
    echo "🔍 Getting task status: $task_id"
    
    curl -s "${BASE_URL}/tasks/$task_id" | jq '.'
}

# Function to get system stats
get_stats() {
    echo "📊 Getting system statistics..."
    curl -s "${BASE_URL}/stats" | jq '.'
}

# Function to get worker status
get_workers() {
    echo "👷 Getting worker status..."
    curl -s "${BASE_URL}/workers" | jq '.'
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "❌ jq is required for this demo. Please install it first."
    exit 1
fi

# Check if server is running
check_server

echo ""
echo "1️⃣ Creating various tasks..."
echo "----------------------------"

# Create email task
email_task=$(create_task "email" '{"recipient": "john@example.com", "subject": "Welcome!", "body": "Hello John!"}' 0)
sleep 1

# Create high-priority image resize task  
image_task=$(create_task "image_resize" '{"image_url": "https://example.com/photo.jpg", "width": 800, "height": 600}' 5)
sleep 1

# Create data processing task
data_task=$(create_task "data_process" '{"data_source": "/path/to/data.csv", "operation": "aggregate"}' 0)
sleep 1

# Create webhook task
webhook_task=$(create_task "webhook" '{"url": "https://api.example.com/callback", "method": "POST", "data": {"event": "task_completed"}}' 2)

echo ""
echo "2️⃣ Checking system status..."
echo "----------------------------"
get_stats
echo ""
get_workers

echo ""
echo "3️⃣ Waiting for tasks to process..."
echo "-----------------------------------"
sleep 8

echo ""
echo "4️⃣ Checking task results..."
echo "----------------------------"

echo "📧 Email Task:"
get_task $email_task
echo ""

echo "🖼️  Image Resize Task:"  
get_task $image_task
echo ""

echo "📊 Data Processing Task:"
get_task $data_task
echo ""

echo "🔗 Webhook Task:"
get_task $webhook_task

echo ""
echo "5️⃣ Final system statistics..."
echo "------------------------------"
get_stats

echo ""
echo "🎉 Demo completed! Check the server logs to see detailed processing information."
echo ""
echo "💡 Try creating more tasks:"
echo "   curl -X POST ${BASE_URL}/tasks \\"
echo "        -H 'Content-Type: application/json' \\"
echo "        -d '{\"type\": \"email\", \"payload\": {\"recipient\": \"test@example.com\", \"subject\": \"Test\"}}'"
echo ""