# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install git and build tools for CGO
RUN apk add --no-cache git build-base

# Copy source code first
COPY . .

# Download dependencies (use default GOPROXY with fallbacks)
RUN go mod download || go mod download
RUN go mod tidy || go mod tidy

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -o nas-dop ./cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/nas-dop .

# Create data directory
RUN mkdir -p /data/db

# Expose port
EXPOSE 8080

# Set default environment variables
ENV ROOT=/data \
    DB_PATH=/data/db/app.sqlite \
    PORT=8080 \
    PUID=0 \
    PGID=0

# Run as non-root user (optional, can be overridden with PUID/PGID)
# USER 1000:1000

ENTRYPOINT ["/app/nas-dop"]
