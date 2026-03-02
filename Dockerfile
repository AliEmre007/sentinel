# Stage 1: The Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache dependencies first (Best Practice for fast builds)
COPY go.mod go.sum ./
RUN go mod download

# Copy your actual source code into the image
COPY . .

# Compile a standalone, statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o sentinel-binary .

# Stage 2: The Production Server
FROM alpine:latest

WORKDIR /root/

# Copy ONLY the compiled binary from Stage 1. Leave the source code behind!
COPY --from=builder /app/sentinel-binary .

# Expose the API port
EXPOSE 8080

# Run the compiled binary (No Air, no live-reloading!)
CMD ["./sentinel-binary"]
