# 1. 前端构建
FROM node:22-alpine AS frontend
WORKDIR /web
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY frontend/ .
RUN pnpm build

# 2. admin 构建（embed 前端）
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -o xxl-job-admin-go ./cmd/server

# 3. 运行
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=backend /app/xxl-job-admin-go .
COPY backend/configs/config.yaml ./configs/
COPY backend/migrations ./migrations
EXPOSE 8080 9090
ENTRYPOINT ["./xxl-job-admin-go"]
