# Step 1: build image
FROM golang:1.21 AS builder

# Cache the dependencies
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

# Compile the application
COPY . /app
RUN --mount=type=cache,target=/root/.cache/go-build ./scripts/build.sh

# Step 2: build the image to be actually run
FROM alpine:3.18.4
RUN mkdir /tmp/kopia-dist && \
    wget -o /tmp/kopia-dist/kopia-0.15.0-linux-x64.tar.gz https://github.com/kopia/kopia/releases/download/v0.15.0/kopia-0.15.0-linux-x64.tar.gz && \
    tar -C /tmp/kopia-dist -xvzf kopia-0.15.0-linux-x64.tar.gz && \
    cp /tmp/kopia-dist/kopia-0.15.0-linux-x64/kopia /usr/bin/kopia && \
    rm -rf /tmp/kopia-dist
USER 10001:10001
COPY --from=builder /app/bin/plugin-pvc-backup /app/bin/plugin-pvc-backup
ENTRYPOINT ["/app/bin/plugin-pvc-backup"]