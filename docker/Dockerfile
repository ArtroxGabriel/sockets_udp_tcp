# Build stage
FROM golang:1.27.1-alpine AS builder

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Copy source code
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

# Compile all binaries with optimizations
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/server_udp ./cmd/server_udp && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/client_udp ./cmd/client_udp && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/server_tcp ./cmd/server_tcp && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/client_tcp ./cmd/client_tcp && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/server_proto ./cmd/server_proto && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/client_proto ./cmd/client_proto && \
    go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /app/bin/experiment ./cmd/experiment

# Runtime stage
FROM alpine:3.21

# iproute2 provides tc for kernel packet loss simulation (netem)
RUN apk add --no-cache \
    iproute2 \
    ca-certificates

WORKDIR /app

# Copy compiled binaries to /usr/local/bin
COPY --from=builder /app/bin/* /usr/local/bin/

# Copy entrypoint script
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8081/udp 8082/tcp 8083/tcp

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["server_udp", "--port", "8081"]
