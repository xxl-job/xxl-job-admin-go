# CLAUDE.md

本文件为 AI 协作（Claude Code 等）提供项目上下文。

## 项目

**xxl-job-admin-go**：XXL-JOB 调度中心的 Go 现代化重实现。
- gRPC/ConnectRPC 长连接架构，执行器↔调度中心单一连接复用注册/心跳/结果/日志/负载
- **无 Redis**：分布式锁用 MySQL 租约表（leader 选举），注册中心纯内存（gRPC 连接态）
- **不兼容** Java xxl-job 协议

## Monorepo 布局

```
backend/     Go 调度中心 + 执行器 SDK（cmd/internal/proto/migrations/configs）
frontend/    vue-admin 前端（Vue3 + Vite + TS + Element-Plus + Pinia）
docs/        design/（架构/API 文档）+ VitePress 文档站源码
deploy/      docker-compose.yml + k8s/
```

## 常用命令

```bash
# 后端（在 backend/ 下执行）
cd backend && make test       # 测试（-race）
cd backend && make lint       # golangci-lint（与 CI 一致）
cd backend && make vet        # go vet 快速检查
cd backend && make proto      # buf lint + generate
cd backend && make build      # 编译到 bin/
cd backend && make up         # docker compose（compose 文件在 deploy/）

# 前端
cd frontend && pnpm install && pnpm build

# 文档
cd docs && pnpm install && pnpm docs:build
```

## 关键约定

- **module 名**：保持 `xxl-job-admin-go`（应用非库，import 路径稳定，proto/buf 零路径耦合）
- **流量泳道**：执行器 `cluster` 标签走 `RegisterRequest.meta["cluster"]`，proto 零改动；任务 `Job.Cluster` 指定目标 cluster
- **响应格式**：`{code:0, data, msg}`（适配 vue-admin），HTTP 统一 200
- **注释语言**：与现有代码库一致（中文）
- **git 约束**：提交身份统一 `如水 <rushui@qq.com>`，绝不暴露真实身份；不擅自 commit/push，需用户明确同意

## 验证（发布前必跑）

```bash
cd backend && gofmt -l . && go vet ./... && go build ./... && go test ./... -race -count=1 && golangci-lint run && buf lint
cd frontend && pnpm build
cd docs && pnpm docs:build
```
