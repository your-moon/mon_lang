# Stage 1: Build frontend
FROM oven/bun:1 AS frontend
WORKDIR /build/web
COPY playground/web/package.json playground/web/bun.lock* ./
RUN bun install --frozen-lockfile
COPY playground/web/ ./
RUN bun run build

# Stage 2: Build Go binaries
FROM golang:1.23 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy frontend build into static dir for embed
COPY --from=frontend /build/web/out /build/playground/static
# Build compiler
RUN CGO_ENABLED=0 go build -o mon_lang .
# Build playground server
RUN cd playground && CGO_ENABLED=0 go build -o playground-server .

# Stage 3: Runtime
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    binutils \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /build/mon_lang ./mon_lang
COPY --from=builder /build/playground/playground-server ./playground-server
COPY --from=builder /build/stdlib ./stdlib
COPY --from=builder /build/playground/examples ./examples

EXPOSE 8080
CMD ["./playground-server", "-compiler", "./mon_lang", "-stdlib", "./stdlib", "-port", "8080"]
