# Build stage
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o storage-api .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the binary from builder
COPY --from=builder /app/storage-api .

ENV STORAGE_ROOT_PATH=/data/storage

# Create storage directory
RUN mkdir -p $STORAGE_ROOT_PATH 

# Expose port
EXPOSE 3000

# Run the application
CMD ["./storage-api"]
