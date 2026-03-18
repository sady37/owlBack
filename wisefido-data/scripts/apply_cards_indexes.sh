#!/bin/bash

# Apply optimized indexes to cards table
# This script should be run after the application is stopped to avoid conflicts

set -e

echo "Applying optimized indexes to cards table..."

# Source environment variables if .env file exists
if [ -f .env ]; then
    source .env
fi

# Set default values if environment variables are not set
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-owlrd}

# Export password for psql
export PGPASSWORD=$DB_PASSWORD

# Apply the indexes
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f optimize_cards_indexes.sql

echo "Indexes applied successfully!"