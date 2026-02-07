# Multi-stage build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o github-discovery ./services/github-discovery

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/
COPY --from=builder /app/github-discovery .
RUN mkdir -p /logs

EXPOSE 8081
CMD ["./github-discovery"]
