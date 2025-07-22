#!/bin/bash

# Build da aplicação S3-Like
echo "🔨 Building S3-Like Docker image..."

# Build da imagem
docker build -t s3like:latest .

if [ $? -eq 0 ]; then
  echo "✅ Build completed successfully!"
  echo "📦 Image: s3like:latest"
else
  echo "❌ Build failed!"
  exit 1
fi
