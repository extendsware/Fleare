# Multi-stage Dockerfile for Fleare Go Application
# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Set build arguments
ARG APP_VERSION
ARG BUILD_DATE

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    make \
    bash

# Set working directory
WORKDIR /app

# Copy go mod and sum files for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Set build date if not provided
RUN if [ -z "$BUILD_DATE" ]; then \
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
    fi

# Build the application with optimization and version info
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o fleare \
    .

# Verify the binary
RUN ./fleare --version

# Stage 2: Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fleare ./

# Copy static files and configuration if they exist
COPY --from=builder /app/scripts/install-docker.sh ./install.sh

RUN chmod +x /app/install.sh

RUN bash /app/install.sh

RUN rm /app/fleare
RUN rm /app/install.sh

RUN tee /app/entrypoint.sh > /dev/null <<'EOL'
#!/bin/sh
exec /usr/local/bin/fleare "$@"
EOL

RUN chmod +x /app/entrypoint.sh

EXPOSE 9219

ENTRYPOINT ["/app/entrypoint.sh"]
