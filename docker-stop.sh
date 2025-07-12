#!/bin/bash

POSTGRES_CONTAINER="s3like-postgres"
APP_CONTAINER="s3like-app"
NETWORK_NAME="s3like-network"

echo "🛑 Stopping S3-Like service..."

# Parar containers
docker stop $APP_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true

# Remover containers
docker rm $APP_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true

# Remover network
docker network rm $NETWORK_NAME 2>/dev/null || true

echo "✅ S3-Like service stopped!"
