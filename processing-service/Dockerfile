# Stage 1: Build the Go application
FROM golang:1.24-bookworm AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker's cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o processing-service ./main.go

# Stage 2: Create a minimal, secure final image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/processing-service ./

# The command to run the application when the container starts
CMD ["/app/processing-service"]