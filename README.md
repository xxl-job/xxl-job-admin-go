# XXL-JOB-Admin-Go

[![CI](https://github.com/xxl-job/xxl-job-admin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/xxl-job/xxl-job-admin-go/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[English](README-en.md) | 中文

XXL-JOB 调度中心的 Go 现代化重实现。**不兼容** Java xxl-job 协议，采用 gRPC/ConnectRPC 长连接架构，无 Redis 依赖。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                   调度中心 xxl-job-admin-go                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │ Gin API  │  │ Scheduler│  │ Registry │  │  Trigger   │  │
│  │ /metrics │  │ 时间轮   │  │ 内存注册 │  │ 路由/阻塞  │  │
│  │ WS 日志  │  │ 5s 预读  │  │          │  │            │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └─────┬──────┘  │
│       │             │             │              │         │
│  ┌────▼─────────────▼─────────────▼──────────────▼──────┐  │
│  │       Service（任务/日志/用户/告警/审计/分区）        │  │
│  └────┬─────────────────────────────────────────────┬───┘  │
│  ┌────▼──────┐                       ┌──────────────▼────┐ │
│  │  MySQL    │  ←租约锁/持久化        │  ExecutorLink     │ │
│  │ (8表+分区)│                       │  gRPC 双向流      │ │
│  └───────────┘                       └──────────────┬────┘ │
└──────────────────────────────────────────────────────┼──────┘
                       ▲  gRPC 长连接（注册/心跳/结果/日志/负载）  │
            ┌──────────┴──────────────────────────────────┴─┐
            │              执行器集群（gRPC SDK）            │
            │   Go 执行器（internal/executor + cmd/sample） │
            └───────────────────────────────────────────────┘
                       ▲  RESTful API + WebSocket
                ┌──────┴──────┐
                │  vue-admin  │  Vue3+Vite+TS+Element-Plus
                └─────────────┘
```

## 设计要点

- **调度与执行分离**：调度中心（admin）+ 分布式执行器（gRPC 长连接）
- **gRPC 双向流**：执行器↔admin 单一长连接复用注册/心跳/结果/日志/负载，下线秒级感知
- **无 Redis**：分布式锁用 MySQL 租约表（`scheduler_lease`，leader 选举），注册中心纯内存（gRPC 连接态）
- **强类型全栈**：Protobuf + buf + ConnectRPC
- **前端**：vue-admin（Vue3 + Vite + TS + Element-Plus + Pinia）
- **运维**：Prometheus metrics、审计日志、告警接口、分区表自动维护、实时日志 WebSocket

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.25 · ConnectRPC · GORM · MySQL 8.0 · viper · zap · golang-jwt · prometheus · gorilla/websocket |
| IDL | Protobuf + buf + protoc-gen-connect-go |
| 前端 | vue-admin（Vue3 + Vite + TS + Element-Plus + Pinia） |
| 调度 | robfig/cron（6 字段含秒）+ 自研时间轮 + 快慢线程池 |

## 目录结构

```
backend/      Go 调度中心 + 执行器 SDK
  cmd/server  cmd/executor-sample
  internal/   conf model store pkg/lock registry scheduler trigger
              rpc api auth alarm metrics logstream executor gen
  proto/xxl/v1/   migrations/   configs/
frontend/     vue-admin（Vue3 + Vite + TS + Element-Plus + Pinia）
docs/         design/（架构/API 文档）+ VitePress 文档站源码
deploy/       docker-compose.yml + k8s/
```

## 快速开始

**demo 模式（一键体验，推荐首次）**：自动 seed 示例任务 + 启动内嵌执行器，启动即有数据可看（任务/执行器/日志/仪表盘）。

```bash
cd deploy && JWT_SECRET="$(openssl rand -hex 32)" docker compose up -d   # 先起 MySQL
cd backend && JWT_SECRET=test-secret-0123456789abcdef \
  DB_DSN="root:root@tcp(127.0.0.1:3306)/xxl_job?charset=utf8mb4&parseTime=True&loc=Local" \
  go run ./cmd/server --demo
```

**完整启动**：

```bash
# 1. 生成 proto 代码（在 backend/）
cd backend && make proto

# 2. 启动 MySQL + admin（在 backend/，compose 文件位于 deploy/）
cd backend && make up

# 3. 启动示例执行器（连 admin，注册 demoJob/slowJob）
cd backend && go run ./cmd/executor-sample -addr http://localhost:9090 -app demo-app

# 4. 启动前端
cd frontend && pnpm install && pnpm dev   # http://localhost:5173
```

登录 `admin / 123456` → 任务管理 → 新建任务（app=demo-app, handler=demoJob, cron=`0/5 * * * * ?`）→ 启动 → 调度日志查看结果，仪表盘看统计。

## API 一览

| 模块 | 端点 |
|---|---|
| 认证 | `POST /api/auth/sessions`、`POST /api/auth/tokens/refresh`、`DELETE /api/auth/sessions`、`GET /api/auth/users/me` |
| 任务 | `/api/jobs` CRUD + `/api/jobs/:id/{start,stop,trigger}` |
| 执行器 | `/api/executors` CRUD |
| 日志 | `GET /api/logs` + `GET /api/logs/:id/stream`（WebSocket 实时） |
| 仪表盘 | `GET /api/dashboard` |
| 审计 | `GET /api/audit` |
| 用户 | `/api/users` CRUD |
| 指标 | `GET /metrics`（Prometheus） |
| gRPC | `ExecutorLink.Link`（双向流）、`AdminService.{TriggerTask,StreamLogs,QueryTaskStatus}` |

## 配置

`configs/config.yaml`，环境变量覆盖（`.` 转 `_`、全大写）：`DB_DSN`、`SERVER_HTTP_ADDR`、`JWT_SECRET` 等。

## 实施阶段

| 阶段 | 内容 | 状态 |
|---|---|---|
| P1 基座 | 骨架、配置、DB、租约锁、gRPC 骨架、JWT、容器化 | ✅ |
| P2 调度核心 | 时间轮 + 5s 预读 + 快慢线程池 + leader 选举 | ✅ |
| P3 RPC 业务 | ExecutorLink 双向流全功能 + 9 种路由 + 阻塞策略 | ✅ |
| P4 前端业务 | vue-admin 集成 + 任务管理页面 + REST API | ✅ |
| P5 运维增强 | metrics / 审计 / 告警 | ✅ |
| P6 执行器 SDK | Go gRPC 执行器 SDK + 示例执行器 | ✅ |
| P7 前端补全 + 实时日志 | 执行器/日志/仪表盘页面 + WebSocket 日志推送 | ✅ |
| P8 分区维护 + 测试 | task_runs 分区自动管理 + 单元测试 | ✅ |
| P9 用户管理 + 文档 | 用户 CRUD + RBAC + 文档完善 | ✅ |
| P10 执行器日志回传 | JobContext.Log→LogChunk→Hub→WS | ✅ |
| P11 部署配置 | 生产配置/K8s manifest/前端环境示例 | ✅ |
| P12 开源治理 | LICENSE/CONTRIBUTING/CI/golangci/模板/docs | ✅ |
| P13 GLUE+任务依赖 | yaegi 在线编程 + 父子任务编排 | ✅ |
| P14 失败重试+取消 | retryJob + CancelTask per-task context | ✅ |
| P15 GLUE 版本管理 | xxl_job_logglue 30 版本回溯 | ✅ |
| P16 脚本任务 | Shell/Python（os/exec） | ✅ |
| P17 accessToken | 调度中心↔执行器认证 | ✅ |
| P18 API 集成测试 | sqlite+httptest，4 包测试全通过 | ✅ |

## 验证状态

- 代码层：`go build` / `go vet` / `go test` / `buf lint` 全通过
- 二进制：`bin/xxl-job-admin-go`（36M）+ `bin/xxl-executor-sample`（12M）
- 前端：`pnpm dev` 启动正常，任务/执行器/日志/审计/用户/仪表盘路由全 HTTP 200
- 端到端运行：受构建环境网络限制（docker hub / github 不可达，本地无 MySQL）未实跑，需在可用环境验证

## 协议

GPL-3.0，与上游 xxl-job 一致（见 [LICENSE](LICENSE)）。
