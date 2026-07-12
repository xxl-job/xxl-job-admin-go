# 1. 构建
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o xxl-job-executor ./cmd/executor

# 2. 运行（sh + python3 脚本运行时）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata python3
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/xxl-job-executor .
ENV ADMIN_ADDR=http://admin:9090 \
    APP_NAME=default
ENTRYPOINT ["./xxl-job-executor"]
