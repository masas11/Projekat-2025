# Build stage
FROM golang:1.21-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy shared module first
COPY shared /shared
WORKDIR /shared
RUN go mod download

# Copy saga-service
WORKDIR /app
COPY saga-service/go.mod saga-service/go.sum* ./
RUN go mod download
RUN go mod tidy

COPY saga-service/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
