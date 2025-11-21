# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory inside the container
WORKDIR /app

# Install any necessary tools
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .
RUN go test ./...
# Build the Go binary
RUN go build -o main .

# Stage 2: Minimal runtime container
FROM alpine:latest

# Set working directory inside the minimal image
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY .env .
# Expose port (adjust if needed)
EXPOSE 8285

# Command to run the executable
CMD ["./main"]