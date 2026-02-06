# Multi-stage build for metadata service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the metadata service
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o metadata ./services/metadata

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs curl

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/metadata .

# Create necessary directories
RUN mkdir -p /data /logs

# Expose port
EXPOSE 8086

# Run the application
CMD ["./metadata"]
