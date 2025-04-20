#!/bin/bash

# Configuration
IMAGE_NAME="kom-mcp-server"
VERSION=$(git describe --tags --always --dirty)

# Use environment variable for registry or default to localhost:5000
REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}

# # Check if we need to log into the registry
# if [ "$REGISTRY" != "localhost:5000" ]; then
#     echo "Using remote registry: $REGISTRY"
#     if [ -z "$DOCKER_USERNAME" ] || [ -z "$DOCKER_PASSWORD" ]; then
#         echo "Error: DOCKER_USERNAME and DOCKER_PASSWORD environment variables are required for remote registry"
#         exit 1
#     fi
#     # echo "Logging into Docker registry..."
#     echo "$DOCKER_PASSWORD" | docker login $REGISTRY -u "$DOCKER_USERNAME" --password-stdin
# fi

# Ensure QEMU is installed for cross-platform builds
docker run --privileged --rm tonistiigi/binfmt --install all

# Remove existing builder if it exists
docker buildx rm multi-arch-builder 2>/dev/null || true

# Create and use a new builder instance
echo "Setting up Docker buildx..."
docker buildx create --name multi-arch-builder --driver docker-container --platform linux/amd64,linux/arm64 --bootstrap || true
docker buildx use multi-arch-builder

# Build and push the multi-architecture Docker image
echo "Building and pushing multi-architecture Docker image ${REGISTRY}/${IMAGE_NAME}:${VERSION}..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --push \
    -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} \
    -t ${REGISTRY}/${IMAGE_NAME}:latest \
    -f build/Dockerfile .

# Verify the manifest
echo "Verifying image manifest..."
docker buildx imagetools inspect ${REGISTRY}/${IMAGE_NAME}:latest

# Clean up the builder
docker buildx rm multi-arch-builder 