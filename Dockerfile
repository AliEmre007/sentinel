# 1. Use the Go compiler image
FROM golang:alpine

# 2. Set our working directory
WORKDIR /app

# 3. Install 'Air' (The Live Reloader)
RUN go install github.com/air-verse/air@latest

# 4. Download our Go modules (like package.json)
COPY go.mod ./
RUN go mod download

# Notice we are NO LONGER copying the main.go file here!
# The Volume Mount will handle that automatically.

# 5. Tell the container to start the Air watcher
CMD ["air"]
