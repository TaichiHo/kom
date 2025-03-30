#!/bin/bash

# Configuration
IMAGE_NAME="kom-mcp-server"
VERSION=$(git describe --tags --always --dirty)
DOCKER_REGISTRY="docker.io/taichiho"  # Docker Hub username

# Build the Docker image
echo "Building Docker image ${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}..."
docker build -t ${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION} -t ${DOCKER_REGISTRY}/${IMAGE_NAME}:latest -f build/Dockerfile .

# Push the Docker image if requested
if [ "$1" = "push" ]; then
    echo "Logging in to Docker Hub..."
    docker login

    echo "Pushing Docker image to Docker Hub..."
    docker push ${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}
    docker push ${DOCKER_REGISTRY}/${IMAGE_NAME}:latest
fi 