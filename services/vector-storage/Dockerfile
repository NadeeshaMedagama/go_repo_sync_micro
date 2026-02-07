# Multi-stage build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vector-storage ./services/vector-storage

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/vector-storage .
RUN mkdir -p /logs
EXPOSE 8084
CMD ["./vector-storage"]
