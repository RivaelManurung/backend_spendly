# Stage 1: Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Use CGO_ENABLED=0 for a static binary that works in alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Final stage
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy migrations and agents (if needed by the app)
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.github/agents ./agents
# Note: In main.go the prompt path is ".github/agents/prompts/". 
# We should ensure the path matches what the binary expects.
# I'll copy the full structure to be safe.
COPY --from=builder /app/.github ./.github

# Expose port
EXPOSE 8080

# Environment variables defaults
ENV APP_ENV=production
ENV HTTP_PORT=8080

# Run the binary
CMD ["./main"]
