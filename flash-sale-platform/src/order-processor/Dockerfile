# Build stage
FROM golang:1.25.1 AS builder

WORKDIR /src

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /order-processor ./cmd/processor

# Runtime stage
FROM alpine:3.19

WORKDIR /src

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /order-processor /src/order-processor

# Create non-root user
RUN adduser -D -g '' srcuser
USER srcuser

# Health check - processor doesn't have HTTP endpoint, so we check process
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep order-processor || exit 1

ENTRYPOINT ["/src/order-processor"]