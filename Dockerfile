# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments
ARG GIT_TAG=""
ARG GIT_COMMIT=""
ARG GIT_BRANCH=""
ARG BUILD_TIME=""
ARG BUILD_NUMBER=""

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s \
    -X 'main.GitTag=${GIT_TAG}' \
    -X 'main.GitCommit=${GIT_COMMIT}' \
    -X 'main.GitBranch=${GIT_BRANCH}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.BuildNumber=${BUILD_NUMBER}'" \
    -o Mastersteam .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/Mastersteam .

# Expose port
EXPOSE 8080

# Set environment variables
ENV STEAM_API_KEY=""

# Run the application
CMD ["./Mastersteam"]

