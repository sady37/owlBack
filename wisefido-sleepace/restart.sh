#!/bin/bash

# Get the current directory name (as the service name)
SERVICE_NAME=$(basename "$PWD")
echo "Service Name: $SERVICE_NAME"

# Define the port the service listens to (adjust per service)
SERVICE_PORT=9005  # Change this port for each service

EXECUTABLE_NAME="main.out"

# Step 1: Find and kill the process listening on the port
PID=$(lsof -iTCP -sTCP:LISTEN -P -n | grep ":$SERVICE_PORT" | awk '{print $2}')

if [ -n "$PID" ]; then
    echo "Stopping service running on port $SERVICE_PORT (PID: $PID)..."
    kill -9 "$PID"
    echo "Service stopped."
else
    echo "No service found running on port $SERVICE_PORT."
fi

# Step 2: Build the service
echo "Building the service '$SERVICE_NAME'..."
if go build -o "$EXECUTABLE_NAME" main.go; then
    echo "Build successful. Executable: $EXECUTABLE_NAME"
else
    echo "Build failed. Exiting."
    exit 1
fi

# Step 3: Start the service
echo "Starting the service '$SERVICE_NAME'..."
nohup ./"$EXECUTABLE_NAME" > "nohup.out" 2>&1 &
NEW_PID=$!
echo "Service restarted with PID: $NEW_PID"

# Step 4: Poll the port to check if the service is listening
TIMEOUT=20  # Maximum wait time in seconds
SLEEP_INTERVAL=2  # Check every 1 second
elapsed=0

while ! lsof -i tcp:$SERVICE_PORT >/dev/null; do
    if [ $elapsed -ge $TIMEOUT ]; then
        echo "Failed to start the service '$SERVICE_NAME' within $TIMEOUT seconds. Check logs."
        exit 1
    fi
    sleep $SLEEP_INTERVAL
    elapsed=$((elapsed + SLEEP_INTERVAL))
done

echo "Service '$SERVICE_NAME' is running on port $SERVICE_PORT."