# Build stage
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

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
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -tags=k8s_1_28 -o mcp-server ./main.go

# Final stage
FROM --platform=$TARGETPLATFORM alpine:3.18

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 10001 mcp

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