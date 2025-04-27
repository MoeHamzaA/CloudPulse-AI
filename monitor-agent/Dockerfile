# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o monitor-agent .

# Final stage
FROM debian:bullseye-slim

# Install CA certificates and security updates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/monitor-agent .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Create non-root user
RUN useradd -m -u 1000 appuser
USER appuser

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./monitor-agent"] 