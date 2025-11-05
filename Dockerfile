# Multi-stage build for optimal image size
FROM golang:1.25-alpine AS builder

WORKDIR /app

# # Copy go mod files
# COPY go.mod go.sum ./

# # Download dependencies
# RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dcn-netbox-infra-check ./cmd/dcn-netbox-infra-check

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the binary from builder
COPY --from=builder /app/dcn-netbox-infra-check .

# Create directories for config and secrets
RUN mkdir -p /app/config /app/secrets

# Run the application
CMD ["./dcn-netbox-infra-check"]
