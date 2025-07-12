#!/bin/bash

# Configurações
POSTGRES_CONTAINER="s3like-postgres"
APP_CONTAINER="s3like-app"
NETWORK_NAME="s3like-network"
STORAGE_PATH="$(pwd)/storage"

echo "🚀 Starting S3-Like service with Docker CLI..."

# Criar network se não existir
echo "📡 Creating Docker network..."
docker network create $NETWORK_NAME 2>/dev/null || echo "Network already exists"

# Parar e remover containers existentes
echo "🛑 Stopping existing containers..."
docker stop $POSTGRES_CONTAINER $APP_CONTAINER 2>/dev/null || true
docker rm $POSTGRES_CONTAINER $APP_CONTAINER 2>/dev/null || true

# Criar diretório de storage
mkdir -p "$STORAGE_PATH"

# Executar PostgreSQL
echo "🗄️  Starting PostgreSQL..."
docker run -d \
  --name $POSTGRES_CONTAINER \
  --network $NETWORK_NAME \
  -e POSTGRES_DB=s3like \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -v s3like_postgres_data:/var/lib/postgresql/data \
  postgres:15-alpine

# Aguardar PostgreSQL inicializar
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Executar aplicação S3-Like
echo "🚀 Starting S3-Like application..."
docker run -d \
  --name $APP_CONTAINER \
  --network $NETWORK_NAME \
  -p 8080:8080 \
  -e DB_HOST=$POSTGRES_CONTAINER \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password \
  -e DB_NAME=s3like \
  -e DB_SSLMODE=disable \
  -e JWT_SECRET=your-super-secret-jwt-key-change-in-production \
  -e STORAGE_PATH=/storage \
  -e SERVER_PORT=8080 \
  -v "$STORAGE_PATH":/storage \
  s3like:latest

if [ $? -eq 0 ]; then
  echo "✅ S3-Like service started successfully!"
  echo "🌐 API available at: http://localhost:8080"
  echo "📊 Health check: http://localhost:8080/health"
  echo ""
  echo "📋 Container status:"
  docker ps --filter "name=s3like"
else
  echo "❌ Failed to start S3-Like service!"
  exit 1
fi
