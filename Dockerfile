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
RUN go build -o main .

# Atlas installation
RUN curl -sSf https://atlasgo.sh | sh

# Final stage for Go app
FROM debian:bullseye-slim AS final

WORKDIR /app

# Update CA certificates
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main
COPY --from=builder /app/atlas.hcl /app/atlas.hcl
COPY --from=builder /usr/local/bin/atlas /usr/local/bin/atlas

# Copy the migrations directory
COPY --from=builder /app/migrations /app/migrations

# Creating a shell script to run migrations and start the app
RUN echo '#!/bin/sh\n\
set -e\n\
/usr/local/bin/atlas migrate hash\n\
/usr/local/bin/atlas migrate apply --env prod --url "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?search_path=public&sslmode=disable"\n\
/app/main' > /app/start.sh
RUN chmod +x /app/start.sh
CMD ["/app/start.sh"]