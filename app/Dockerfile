# ----------- Stage 1: Build the Go binary -----------
FROM golang:1.25-alpine AS builder

# Install any dependencies (like git if needed)
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first (for better build caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the binary (static, no CGO)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

# ----------- Stage 2: Create minimal runtime image -----------
FROM alpine:3.20

# Create a non-root user (recommended for security)
RUN adduser -D appuser

# Set the working directory
WORKDIR /home/appuser

# Copy binary from builder stage
COPY --from=builder /app/app .

# Set permissions
RUN chown appuser:appuser ./app
USER appuser

# Expose the port your app runs on
EXPOSE 8080

# Command to run your app
CMD ["./app"]
