#!/bin/bash

# Configuration
IMAGE_NAME="kom-mcp-server"
VERSION=$(git describe --tags --always --dirty)

# Use environment variable for registry or default to localhost:5000
REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}

# Check if we need to log into the registry
if [ "$REGISTRY" != "localhost:5000" ]; then
    echo "Using remote registry: $REGISTRY"
    if [ -z "$DOCKER_USERNAME" ] || [ -z "$DOCKER_PASSWORD" ]; then
        echo "Error: DOCKER_USERNAME and DOCKER_PASSWORD environment variables are required for remote registry"
        exit 1
    fi
    echo "Logging into Docker registry..."
    echo "$DOCKER_PASSWORD" | docker login $REGISTRY -u "$DOCKER_USERNAME" --password-stdin
fi

# Build the Docker image
echo "Building Docker image ${REGISTRY}/${IMAGE_NAME}:${VERSION}..."
docker build -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} -t ${REGISTRY}/${IMAGE_NAME}:latest -f build/Dockerfile .

# Push the Docker image to registry
echo "Pushing Docker image to registry..."
docker push ${REGISTRY}/${IMAGE_NAME}:${VERSION}
docker push ${REGISTRY}/${IMAGE_NAME}:latest 