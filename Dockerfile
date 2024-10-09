# Build stage
FROM golang:latest AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Copy the entire project
COPY . .

# Download dependencies
RUN go mod download

# Build the binary
WORKDIR /app
RUN go build -o main .

# Final stage for Go app
FROM debian:bullseye-slim AS final

WORKDIR /app

# Update CA certificates
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main

# Copy the .env file to the working directory (optional if needed in the container)
COPY .env /app/.env

# Make the binary executable
RUN chmod +x /app/main

# Set the entry point to run the Go binary
ENTRYPOINT ["/app/main"]
