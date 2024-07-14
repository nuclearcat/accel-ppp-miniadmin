# golang
FROM golang:1.22.5-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# accel-miniadmin.go
COPY accel-miniadmin.go .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app
RUN go build -o /go/bin/accel-miniadmin accel-miniadmin.go

# Start a new stage from scratch
FROM alpine:3.20

# Install dependencies
COPY --from=builder /go/bin/accel-miniadmin /usr/bin/accel-miniadmin

WORKDIR /app

COPY static/* /app/static/

# Run the binary program produced by `go build`
CMD ["/usr/bin/accel-miniadmin"]
