# 贡献指南

感谢你对 xxl-job-admin-go 的关注！欢迎贡献代码、提交问题或建议。

## 开发环境

- Go 1.25+
- Node.js 22+ / pnpm 9+
- MySQL 8.0（本地或 Docker）
- buf（proto 代码生成）：`go install github.com/bufbuild/buf/cmd/buf@latest`

## 快速开始

```bash
# 克隆
git clone <repo-url> && cd xxl-job-admin-go

# 后端：依赖 + proto + 启动（在 backend/）
cd backend
go mod tidy
make proto
make up          # docker compose（compose 文件位于 ../deploy/）
make run

# 启动示例执行器
go run ./cmd/executor-sample -addr http://localhost:9090 -app demo-app

# 启动前端
cd ../frontend && pnpm install && pnpm dev
```

## 开发规范

### 代码风格

- 遵循 [Go Code Review](https://go.dev/wiki/CodeReview) 与 [Effective Go](https://go.dev/doc/effective_go)
- 注释与现有代码库语言一致（中文）
- 遵循 SOLID / KISS / DRY / YAGNI 原则
- 接口优先（依赖倒置），策略模式避免 if-else 蔓延

### 提交规范（Conventional Commits）

```
<type>(<scope>): <subject>

type: feat | fix | docs | style | refactor | test | chore | perf
scope: scheduler | trigger | rpc | api | executor | web | ...
```

示例：`feat(scheduler): 支持任务依赖编排`

### PR 流程

1. Fork 仓库并创建特性分支：`feat/xxx`、`fix/xxx`
2. 后端在 `backend/` 下通过 `go build ./...`、`go vet ./...`、`go test ./...`、`buf lint`；前端 `cd frontend && pnpm build`
3. 新增功能需配套测试
4. 提交 PR，描述变更与动机，关联相关 Issue

### 架构约定

- **调度中心**（admin）：调度核心 + RPC + REST API + 前端
- **执行器**：通过 gRPC 双向流（`ExecutorLink`）接入
- **无 Redis**：MySQL 租约锁 + 内存注册中心
- 修改 proto 后须 `make proto` 重新生成代码
- 前端响应格式：`{code:0, data, msg}`（适配 vue-admin）

## 测试

```bash
go test ./... -race -count=1
```

新增包应附单元测试（纯函数/内存逻辑优先，DB 依赖用接口抽象便于 mock）。

## 问题与建议

- Bug 请提 [Issue](../../issues)，附复现步骤与日志
- 功能建议请提 Issue 并标注 `feature request`
- 安全漏洞请按 [SECURITY.md](./SECURITY.md) 私密报告
