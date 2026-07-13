# 镜像部署

xxl-job-admin-go 提供两个官方镜像，开箱即用。

## 镜像清单

| 镜像 | 说明 |
|------|------|
| `xxljob/xxl-job-admin` | 调度中心，前后端一体（go:embed），单端口 8080 |
| `xxljob/xxl-job-executor` | 通用执行器，内置 httpCall/shellExec/mysqlQuery/sendWebhook + Shell/Python 脚本运行时 |

## 一键部署（docker compose）

```bash
JWT_SECRET=your-strong-secret docker compose -f deploy/docker-compose.yml up
```

浏览器访问 `http://localhost:8080`（默认 admin/123456），executor 自动注册（app=default）。

## 一键体验（all-in-one，2 容器）

最少组件体验：mysql + admin（`-executor` 内嵌通用执行器），无需单独部署 executor 容器。

```bash
JWT_SECRET=your-strong-secret \
  docker compose -f deploy/docker-compose.allinone.yml up
```

admin 进程内置通用执行器（app=default），自动注册 `httpCall`/`shellExec`/`mysqlQuery`/`sendWebhook` 四个通用 handler，建 BEAN 任务填 handler 名 + JSON 参数即可调度。仅映射 8080 端口（gRPC 9090 仅容器内本机连接，不对外）。

> 适用：试用、演示、本地评估。**生产请用上面的 3 容器 compose**（executor 独立，便于伸缩与隔离）。

## 单容器运行

### admin

```bash
docker run -d --name admin \
  -e DB_DSN="root:root@tcp(mysql:3306)/xxl_job?charset=utf8mb4&parseTime=True&loc=Local" \
  -e JWT_SECRET="your-strong-secret" \
  -p 8080:8080 -p 9090:9090 \
  xxljob/xxl-job-admin
```

### executor

```bash
docker run -d --name executor \
  -e ADMIN_ADDR=http://admin:9090 \
  -e APP_NAME=my-app \
  xxljob/xxl-job-executor
```

## admin 环境变量

| 变量 | 说明 |
|------|------|
| `DB_DSN` | MySQL DSN（必填） |
| `JWT_SECRET` | JWT 密钥（必填，≥16 字节，非占位符） |
| `SERVER_HTTP_ADDR` | HTTP 端口，默认 :8080 |
| `SERVER_GRPC_ADDR` | gRPC 端口，默认 :9090 |
| `LOG_RETAIN_DAYS` | task_runs 保留天数，默认 30，0=不清理 |

## executor 环境变量

| 变量 | 说明 |
|------|------|
| `ADMIN_ADDR` | admin gRPC 地址（如 http://admin:9090） |
| `APP_NAME` | 执行器 app 名称 |
| `ACCESS_TOKEN` | 调度中心 access_token（可空） |

## 内置 handler

executor 镜像内置 4 个通用 handler，admin 建 BEAN 任务 JobHandler 填名、ExecutorParam 填 JSON 即调度：

> **成功判定**：`httpCall`/`sendWebhook` 按响应状态码判定——2xx 算成功（任务码 200），其余判失败（任务码 500，触发重试/告警）。即 `201/204` 等正常响应也算成功。

### httpCall
```json
{"method":"POST","url":"https://api.example.com","headers":{"Authorization":"Bearer xxx"},"body":"{}","timeout":10}
```

### shellExec
```json
{"cmd":"echo hello && date","timeout":60}
```

### mysqlQuery
```json
{"dsn":"user:pass@tcp(host:3306)/db","sql":"SELECT id,name FROM users LIMIT 10"}
```

### sendWebhook
```json
{"url":"https://oapi.dingtalk.com/robot/send?access_token=xxx","json":{"msgtype":"text","text":{"content":"hello"}}}
```

## 脚本任务

executor 镜像含 sh + python3，admin 建 GLUE 脚本任务（Shell/Python）可直接执行。
