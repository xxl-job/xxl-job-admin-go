# 运行阶段：使用本机预编译的 linux/arm64 二进制（避免容器内 go build 在低内存 VM OOM）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY bin/xxl-job-admin-go .
COPY configs/config.yaml ./configs/config.yaml
COPY migrations ./migrations
EXPOSE 8080 9090
ENTRYPOINT ["./xxl-job-admin-go"]
