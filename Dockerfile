# Step 1: build image
FROM golang:1.21 AS builder

# Cache the dependencies
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

# Compile the application
COPY . /app
RUN ./scripts/build.sh

# Step 2: build the image to be actually run
FROM alpine:3.18.4
USER 10001:10001
COPY --from=builder /app/bin/plugin-pvc-backup /app/bin/plugin-pvc-backup
ENTRYPOINT ["/app/bin/plugin-pvc-backup"]