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

# Install curl for healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# Create app directory
WORKDIR /app

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the pre-built binary
COPY --from=builder /app/main .

# Change ownership of the app directory
RUN chown -R appuser:appgroup /app

# Change to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]