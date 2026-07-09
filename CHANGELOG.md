# Changelog

本项目变更记录遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added
- GLUE 在线编程（yaegi 解释器，执行器侧支持在线 Go 代码）
- 任务依赖编排（父任务成功后触发子任务）
- 失败重试（`executor_fail_retry_count`，内存计数跟踪，成功清除）
- CancelTask 真实取消（执行器 per-task context，收到 Cancel 调 cancel；调度侧 cover-early 阻塞策略下发 CancelTask 取消旧实例，Trigger 维护 logID→conn 映射）
- GLUE 版本管理（`xxl_job_logglue` 表，保留最近 30 版本，`GET /api/jobs/:id/glue-versions`）
- 脚本任务（`glue_type` 2=Shell / 3=Python，os/exec 执行，输出经 jc.Log 回传）
- accessToken 认证（调度中心↔执行器，HTTP middleware 校验 + executor interceptor 注入）
- API 集成测试（sqlite 内存库 + httptest，覆盖登录/鉴权/任务CRUD/用户）
- 开源治理：LICENSE(GPLv3)、CONTRIBUTING、CODE_OF_CONDUCT、SECURITY
- GitHub Actions CI（test/build/lint/buf）、Issue/PR 模板、golangci-lint 配置
- 项目文档（架构/API/部署/执行器开发指南）+ README badge
- RBAC 授权中间件（写操作与敏感查询仅管理员，普通用户返回 403）
- JWT 密钥启动强校验（拒绝空/过短/占位符 `change-me-in-production`）
- API 层 GORM 错误处理 + 分页 helper（`parsePagination`，消除重复）
- `handleTaskResult` 槽位泄漏修复（task_run 查询失败时按 logID 释放阻塞槽位）
- Prometheus 死指标接入（`JobRunning`/`ExecutorOnline` 实际更新）
- WebSocket 同源校验（CSWSH 防护，`WS_ALLOWED_ORIGINS` 白名单）
- GLUE/脚本任务默认禁用（`executor.WithScriptTask(true)` 显式开启）

## [0.1.0] - 2026-07-07

### Added
- **P1 基座**：项目骨架、配置(viper)、GORM 模型、MySQL schema（7 表含分区）、租约锁（`scheduler_lease`，无 Redis）、ConnectRPC 服务骨架、JWT 认证、Dockerfile/docker-compose、vue-admin 前端就位
- **P2 调度核心**：时间轮 + 5s 预读 + 快慢线程池（>500ms 降级 slow pool）+ leader 选举（MySQL 租约）
- **P3 RPC 业务**：ExecutorLink 双向流全功能（register/heartbeat/task_result/log/load）+ 9 种路由策略（First/Last/Round/Random/ConsistentHash/LFU/LRU/Failover/Busyover）+ ShardingBroadcast + 阻塞策略（serial/discard-later/cover-early）+ BlockTracker
- **P4 前端业务**：vue-admin 集成（响应格式 `{code:0,data,msg}` 适配）、任务管理页面、REST API（/api/auth/sessions 双 token、/api/jobs CRUD）
- **P5 运维增强**：Prometheus metrics（trigger_total/fail/running/executor_online）、审计 middleware、告警接口（LogNotifier/EmailNotifier 可扩展）
- **P6 执行器 SDK**：Go gRPC 执行器 SDK（`internal/executor`）+ 示例执行器（`cmd/executor-sample`），断线重连
- **P7 前端补全 + 实时日志**：执行器/日志/仪表盘页面 + WebSocket 实时日志推送（logstream.Hub + `/api/logs/:id/stream`）
- **P8 分区维护 + 测试**：task_runs 分区自动管理（REORGANIZE p_max + DROP 老分区，仅 leader）+ 单元测试（registry/scheduler/trigger）
- **P9 用户管理 + 文档**：用户 CRUD + RBAC（bcrypt，禁删 id=1）+ README 完善
- **P10 执行器日志回传**：JobContext.Log 回调，handler 执行时 LogChunk 回传 → Hub → 前端 WebSocket，实时日志端到端
- **P11 部署配置**：生产配置示例、K8s manifest（多副本 HA）、前端环境示例

### Fixed
- `store.Seed` 初始化 `scheduler_lease` id=1 行（AutoMigrate 只建表不插数据导致 leader 抢不到）
- admin rpcServer 用 h2c 支持 gRPC HTTP/2（标准 http.Server 仅 HTTP/1.1 报 505）
- 执行器客户端用 h2c transport（`http2.Transport{AllowHTTP:true}`，http.DefaultClient 不支持 cleartext h2c）
- `scheduler.fireSecond` 触发 `sec<=now` 所有过期任务（非仅当前秒），`next_time` 基于 `time.Now()` 避免反复触发
- partition 维护检查 `PARTITION_NAME IS NOT NULL`（非分区表在 information_schema 也有 1 条 NULL 记录）
