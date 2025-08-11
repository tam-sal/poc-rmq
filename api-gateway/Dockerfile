# Stage 1: Build the Go application
# Use a full-featured Go image for the build environment
FROM golang:1.24-bookworm AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker's cache
COPY go.mod go.sum ./
# Download all dependencies. This step is only re-run if go.mod or go.sum change.
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary
# Use a static build to avoid runtime dependencies in the final image
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o api-gateway ./main.go

# Stage 2: Create a minimal, secure final image
# Use a distroless image for the final runtime
FROM gcr.io/distroless/static:nonroot

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/api-gateway /app/

# Expose the port the application will run on
EXPOSE 8080

# The command to run the application when the container starts
CMD ["/app/api-gateway"]