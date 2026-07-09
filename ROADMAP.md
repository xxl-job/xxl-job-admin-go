# 路线图

本文档描述 xxl-job-admin-go 的版本规划与演进方向。

## v1.0（当前 · 开源首版）

- **目录规范**：`backend/` + `frontend/` + `docs/` + `deploy/` monorepo 布局
- **代码达标**：RBAC 授权、JWT 密钥校验、API 错误处理、槽位泄漏修复、指标接入、WS 同源校验、脚本开关
- **流量泳道**：执行器 `cluster` 标签路由（借鉴 zeus 概念，复用 `RegisterRequest.meta`，proto 零改动）
- **治理文件**：CODEOWNERS、dependabot、PR 模板、CHANGELOG、SECURITY

## v1.x（服务治理 · 自研轻量）

> zeus 框架经评估不引入（调度系统专用架构，通用微服务框架收益有限），缺失能力自研。

- **熔断**：自研计数器熔断（≈50 行，参考 hystrix/zeus counter），按 `app` 维度
- **限流**：基于 `golang.org/x/time/rate`
- **链路追踪**：OpenTelemetry Go SDK
- **泳道完整隔离**：熔断/限流按 `(app, cluster)` 维度
- **refresh token 独立 + logout 吊销**：接入 session 存储（当前为已知限制）
- **配置热更新**：viper + fsnotify

## v2.0（演进）

- 多调度器协同优化（分片调度、跨集群）
- 更丰富告警渠道（钉钉 / 飞书 / 通用 webhook）
- 执行器多语言 SDK（Java / Python）
- Webhook / 事件触发器
- 任务依赖 DAG 可视化编排

## 反向不兼容变更

遵循 [语义化版本](https://semver.org/lang/zh-CN/)：API/proto 不兼容变更升主版本号。
