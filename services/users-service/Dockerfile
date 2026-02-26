# Build stage
FROM golang:1.24-alpine AS builder

# Install git for go mod tidy
RUN apk add --no-cache git

WORKDIR /app

# Copy shared module first
COPY shared /shared

# Generate go.sum for shared module (MUST be done before building service)
WORKDIR /shared
RUN go mod download
RUN go mod tidy
WORKDIR /app

# Copy users-service go mod files
COPY users-service/go.mod users-service/go.sum* ./
RUN go mod download

# Copy users-service source code
COPY users-service/ ./

# Update go.mod and go.sum if needed
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8001

CMD ["./main"]
