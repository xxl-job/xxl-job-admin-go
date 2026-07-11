# 路线图

本文档描述 xxl-job-admin-go 的版本规划与演进方向。

## v1.0（当前 · 开源首版）

**基础与规范**
- 目录规范：`backend/` + `frontend/` + `docs/` + `deploy/` monorepo
- 代码达标：RBAC 授权、JWT 强密钥校验、API 错误处理、槽位泄漏修复、WS 同源校验、GLUE 脚本开关
- 治理文件：CODEOWNERS、dependabot、PR 模板、CHANGELOG、SECURITY、ROADMAP、MAINTAINERS

**调度核心**
- gRPC 长连接调度（执行器↔调度中心单一连接复用注册/心跳/结果/日志/负载）
- 无 Redis：MySQL 租约锁 leader 选举 + 内存注册中心
- 流量泳道：执行器 `cluster` 标签路由（复用 RegisterRequest.meta，proto 零改动）
- 9 种路由策略 + 3 种阻塞策略 + GLUE 在线代码（Shell/Python/Go）+ 子任务链 + 失败重试
- 任务超时强杀（执行器侧 context cancel）

**可靠性**
- 触发幂等：`job_trigger_dedup` 防 leader 切换窗口重复触发
- 任务按创建者隔离：普通用户只见/管自己创建的，admin 全部

**可观测与运维**
- Prometheus 指标（4 个：trigger/fail/running/executor_online）+ Grafana 大盘模板
- zap 结构化日志 + 审计日志 + 告警通道（钉钉/企微/飞书 webhook + 邮件 SMTP）
- 调度配置热生效（改 CRON 5s 内生效，无需 stop/start）
- 健康检查 `/healthz` + 数据分区维护

**开箱体验**
- 示例数据 + 场景 demo（`backend/examples/`：9 示例任务 + GLUE 脚本）
- 核心单测覆盖（auth 91% / alarm 74% / logstream 100% / registry 97% / conf 85%）
- VitePress 文档站（运维可观测 + 场景示例 + 部署指南）

## v1.x（服务治理 · 自研轻量）

> zeus 框架经评估不引入（调度系统专用架构，通用微服务框架收益有限），缺失能力自研。

- ~~**熔断**：自研计数器熔断，按 `app` 维度~~ ✅ 已完成（`circuit.Breaker`）
- ~~**限流**：基于 `golang.org/x/time/rate`~~ ✅ 已完成（`limiter.Limiter`）
- **链路追踪**：OpenTelemetry Go SDK
- **泳道完整隔离**：熔断/限流按 `(app, cluster)` 维度
- **refresh token 独立 + logout 吊销**：接入 session 存储（当前为已知限制）
- **OpenAPI/Swagger 规范**：gin-swagger 自动生成
- **国际化**：英文 README + 后端错误消息 i18n
- **集成测试**：testcontainers 覆盖 lock/rpc/store（MySQL 容器）

## v2.0（演进）

- 多调度器协同（分片调度、跨集群）
- 执行器多语言 SDK（Java / Python）
- Webhook / 事件触发器
- 任务依赖 DAG 可视化编排

## 反向不兼容变更

遵循 [语义化版本](https://semver.org/lang/zh-CN/)：API/proto 不兼容变更升主版本号。
