# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o governance-alerts-cosmos .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S governance && \
    adduser -u 1001 -S governance -G governance

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/governance-alerts-cosmos .

# Copy configuration
COPY --from=builder /app/config ./config

# Change ownership to non-root user
RUN chown -R governance:governance /app

# Switch to non-root user
USER governance

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
CMD ["./governance-alerts-cosmos", "--config", "config/config.yaml"] 
