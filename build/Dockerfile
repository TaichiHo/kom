# Build stage
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Set GOTOOLCHAIN to auto
ENV GOTOOLCHAIN=auto

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with build tags
RUN CGO_ENABLED=0 GOOS=linux go build -tags=k8s_1_32 -o mcp-server ./main.go

# Final stage
FROM ubuntu:24.04

# Install necessary runtime dependencies
RUN apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -u 10001 -m mcp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mcp-server .

# Set ownership
RUN chown -R mcp:mcp /app

# Use non-root user
USER mcp

# Expose MCP server port
EXPOSE 9096

# Set environment variables
ENV TZ=UTC

# Run the MCP server
ENTRYPOINT ["./mcp-server"] 