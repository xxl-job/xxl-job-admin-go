# XXL-JOB-Admin-Go

[![CI](https://github.com/xxl-job/xxl-job-admin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/xxl-job/xxl-job-admin-go/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[中文](README.md) | **English**

A modern Go reimplementation of the XXL-JOB scheduling center. **Not compatible** with the Java xxl-job protocol — uses a gRPC/ConnectRPC long-connection architecture with no Redis dependency.

## Design Highlights

- **Schedule/execute separation**: admin center + distributed executors (gRPC long-connection)
- **gRPC bidirectional stream**: a single long connection multiplexes register/heartbeat/result/log/load; executor offline detected in seconds
- **No Redis**: distributed lock via a MySQL lease table (`scheduler_lease`, leader election); in-memory registry backed by gRPC connection state
- **Strongly-typed stack**: Protobuf + buf + ConnectRPC
- **Frontend**: vue-admin (Vue3 + Vite + TS + Element-Plus + Pinia)
- **Operations**: Prometheus metrics, audit log, alarm channels, partition maintenance, real-time log WebSocket

## Tech Stack

| Layer | Choice |
|---|---|
| Backend | Go 1.25 · ConnectRPC · GORM · MySQL 8.0 · viper · zap · golang-jwt · prometheus |
| IDL | Protobuf + buf + protoc-gen-connect-go |
| Frontend | vue-admin (Vue3 + Vite + TS + Element-Plus + Pinia) |
| Scheduling | robfig/cron (6-field with seconds) + custom time wheel + fast/slow worker pools |

## Features

- 9 routing strategies + 3 blocking strategies + GLUE online code (Shell/Python/Go) + child tasks + failure retry + timeout kill
- Traffic lane (cluster tag routing)
- **Circuit breaker + rate limiter** (per-app, self-implemented)
- **Trigger idempotency** (safe across leader switch)
- User isolation (per-creator; admin sees all)
- Alarm channels: DingTalk / WeCom / Feishu webhook + Email SMTP
- Hot reload of schedule config (CRON change effective within 5s, no stop/start)

## Directory

```
backend/      Go admin + executor SDK (cmd/internal/proto/migrations/configs)
frontend/     vue-admin (Vue3 + Vite + TS + Element-Plus + Pinia)
docs/         VitePress docs site + design docs
deploy/       docker-compose.yml + k8s/ + grafana/
```

## Quick Start

**Demo mode (one-click, recommended for first run)**: seeds sample jobs + starts an embedded executor — data is ready on startup (jobs/executors/logs/dashboard all populated).

```bash
cd deploy && JWT_SECRET="$(openssl rand -hex 32)" docker compose up -d   # MySQL first
cd backend && JWT_SECRET=test-secret-0123456789abcdef \
  DB_DSN="root:root@tcp(127.0.0.1:3306)/xxl_job?charset=utf8mb4&parseTime=True&loc=Local" \
  go run ./cmd/server --demo
```

**Full setup**:

```bash
# 1. Start MySQL + admin (compose file at deploy/)
export JWT_SECRET="$(openssl rand -hex 32)"
cd deploy && docker compose up -d

# 2. Start sample executor (connects to admin, registers demoJob/slowJob)
cd backend && go run ./cmd/executor-sample -app demo-app

# 3. Start frontend
cd frontend && pnpm install && pnpm dev   # http://localhost:5173
```

Login `admin / 123456` → Job Management → create job (app=`demo-app`, handler=`demoJob`, cron=`0/5 * * * * ?`) → start → view results in schedule logs.

> One-line demo data import: `mysql < backend/examples/seed-demo.sql` (9 sample tasks + GLUE scripts).

## API

Response format: `{"code":0,"data":<T>,"msg":"ok"}` (code 0 = success, HTTP always 200).

| Module | Endpoint |
|---|---|
| Auth | `POST /api/auth/sessions`, `POST /api/auth/tokens/refresh`, `DELETE /api/auth/sessions`, `GET /api/auth/users/me` |
| Jobs | `/api/jobs` CRUD + `/api/jobs/:id/{start,stop,trigger}` |
| Executors | `/api/executors` CRUD |
| Logs | `GET /api/logs` + `GET /api/logs/:id/stream` (WebSocket real-time) |
| Dashboard | `GET /api/dashboard` |
| Metrics | `GET /metrics` (Prometheus) |
| gRPC | `ExecutorLink.Link` (bidi stream), `AdminService.{TriggerTask,QueryTaskStatus}` |

## Testing

- **Unit**: auth 91% / alarm 74% / logstream 100% / circuit 90% / registry 97% / conf 85%
- **Integration**: lock lease failover + rpc handleTaskResult (testcontainers MySQL)
- **E2E**: Playwright (login flow, full stack)
- **CI**: backend / integration / e2e / proto / lint / frontend jobs

## Documentation

Full docs at [docs.cncf.vip/xxl-job/](http://docs.cncf.vip/xxl-job/) (Chinese) — architecture, deployment, operations, API reference, executor development.

## License

GPL-3.0, consistent with upstream xxl-job (see [LICENSE](LICENSE)).
