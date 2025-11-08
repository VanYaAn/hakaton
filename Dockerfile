FROM golang:latest AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bot

FROM alpine:latest

# Установите curl для healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root/

# Copy the pre-built binary
COPY --from=builder /app/main .

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Change to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]