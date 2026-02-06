# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task-assignment-engine .

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for healthcheck
RUN apk --no-cache add ca-certificates curl

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/task-assignment-engine .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./task-assignment-engine"]
