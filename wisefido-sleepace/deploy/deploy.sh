#!/bin/bash

# Ensure two arguments are provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <target> <remote>"
    echo "Example: $0 test wisefido01  OR  $0 dev local"
    exit 1
fi

# Get the arguments
TARGET=$1   # "dev" or "test" or "prod"
REMOTE=$2   # Remote SSH alias (e.g., "wisefido01", "local")

# Configuration
REMOTE_PATH="/var/wisefido/backend/services/sleepace/" # Path to the directory on the remote server
SUPERVISOR_TASK="wisefido-sleepace"      # Supervisor task name
BUILD_TARGET="./build/sleepace"            # Local path to your executable after build
case $TARGET in
    "test")
        CONFIG_FILE="./sleepace-test.yaml"
        ;;
    "dev")
        CONFIG_FILE="./sleepace-dev.yaml"
        ;;
    *)
        echo "Invalid target: $TARGET"
        exit 1
        ;;
esac

# Step 1: Build the Go application
echo "Building the application for $TARGET..."
go build -o $BUILD_TARGET || { echo "Build failed"; exit 1; }

# Step 2: Stop the Supervisor task on the remote server
echo "Stopping the Supervisor task on the remote server ($REMOTE)..."
ssh $REMOTE "sudo supervisorctl stop $SUPERVISOR_TASK" || { echo "Failed to stop Supervisor task"; exit 1; }

# Step 3: Copy files to the remote server
echo "Copying executable and configuration files to the remote server ($REMOTE)..."
scp $BUILD_TARGET $REMOTE:$REMOTE_PATH || { echo "Failed to copy executable"; exit 1; }
scp $CONFIG_FILE $REMOTE:$REMOTE_PATH || { echo "Failed to copy configuration file"; exit 1; }

# Step 4: Start the Supervisor task on the remote server
echo "Starting the Supervisor task on the remote server ($REMOTE)..."
ssh $REMOTE "sudo supervisorctl start $SUPERVISOR_TASK" || { echo "Failed to start Supervisor task"; exit 1; }

echo "Deployment to $TARGET on $REMOTE complete!"
