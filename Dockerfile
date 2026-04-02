# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o spendly-api main.go

# Run stage
FROM alpine:3.19

WORKDIR /app

# Copy binary and configuration
COPY --from=builder /app/spendly-api .
COPY --from=builder /app/.github/agents/prompts .github/agents/prompts
# Copy migrations if needed for entrypoint
COPY --from=builder /app/migrations migrations

# Expose port
EXPOSE 8080

# Command to run
CMD ["./spendly-api"]
