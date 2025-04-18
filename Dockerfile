# Build stage
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create and set working directory
WORKDIR /app

# Install required build tools and dependencies
RUN apk add --no-cache git make

# Copy Go module files
COPY go.mod ./
# Download dependencies and generate go.sum
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" -o /go/bin/libgo-server ./cmd/server

# Final stage
FROM alpine:3.19

# Set label according to OCI image spec
LABEL org.opencontainers.image.title="KVM VM Management API"
LABEL org.opencontainers.image.description="RESTful API for managing KVM virtual machines"
LABEL org.opencontainers.image.url="https://github.com/wroersma/libgo"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.licenses="MIT"

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates libvirt-client libvirt

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create configuration directory
RUN mkdir -p /etc/libgo && chown -R appuser:appgroup /etc/libgo

# Copy the application binary from the builder stage
COPY --from=builder /go/bin/libgo-server /usr/local/bin/

# Copy default configuration
COPY configs/config.yaml.example /etc/libgo/config.yaml

# Use the non-root user
USER appuser

# Expose the API port
EXPOSE 8080

# Set up a health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD [ "wget", "-q", "-O", "-", "http://localhost:8080/health" ]

# Set the default configuration path
ENV CONFIG_PATH=/etc/libgo/config.yaml

# Run the application
ENTRYPOINT ["/usr/local/bin/libgo-server"]
