# Start from the official Golang base image
FROM golang:1.22

# Install air
RUN go install github.com/air-verse/air@latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Ginkgo and Gomega (for testing)
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest
RUN go install github.com/onsi/gomega/...

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
WORKDIR /app/transaction/cmd

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["air"]