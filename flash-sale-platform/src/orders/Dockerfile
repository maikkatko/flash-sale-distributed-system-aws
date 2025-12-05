# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application for a Linux environment, statically linked.
# The -tags=products flag ensures the correct main function is built.
RUN CGO_ENABLED=0 GOOS=linux go build -tags=products -a -o products-api .

# Stage 2: Create the final, minimal image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/products-api .
EXPOSE 8081
CMD ["./products-api"]