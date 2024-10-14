# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set environment variables for Go
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code to the working directory
COPY . .

# Build the Go application
RUN go build -o main .

# Stage 2: Create a lightweight container to run the app
FROM alpine:latest

ENV AZURITE_CONNECTION_STRING=""

# Install necessary certificates
# RUN apk --no-cache add ca-certificates

# Set the working directory for the final container
WORKDIR /root/

# Copy the Go binary from the builder stage
COPY --from=builder /app/main .

# Expose the application port (if necessary)
EXPOSE 8080

# Command to run the Go application
CMD ["./main"]