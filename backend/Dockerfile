# ---------- Stage 1: Build ----------
FROM golang:1.25-alpine AS builder

# Use Chinese mirrors for Alpine APK and Go modules.
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories
RUN apk add --no-cache git ca-certificates tzdata

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server

# ---------- Stage 2: Runtime ----------
FROM alpine:3.21

RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/configs /app/configs

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/swagger/index.html || exit 1

ENTRYPOINT ["/app/server"]
CMD ["-mode", "all"]
