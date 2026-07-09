# 架构设计

## 总体架构

调度与执行分离：调度中心（admin）集群 + 分布式执行器（gRPC 长连接）+ MySQL（持久化与协调）。

```
┌─────────────────────────────────────────────────────────────┐
│                   调度中心 xxl-job-admin-go                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │ Gin API  │  │ Scheduler│  │ Registry │  │  Trigger   │  │
│  │ /metrics │  │ 时间轮   │  │ 内存注册 │  │ 路由/阻塞  │  │
│  │ WS 日志  │  │ 5s 预读  │  │          │  │            │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └─────┬──────┘  │
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
            └───────────────────────────────────────────────┘
```

## 核心组件

### 调度核心（scheduler）

- **时间轮 + 5s 预读**：每 5 秒预读 `jobs` 表 `trigger_next_time <= now+5s` 的任务入内存时间轮；每秒触发 `sec<=now` 的所有任务（含过期补偿）
- **快慢线程池**：fast（200 worker）/ slow（100 worker），任务执行 >500ms 自动降级到 slow pool，避免耗尽调度线程
- **leader 选举**：MySQL 租约表 `scheduler_lease`，`UPDATE ... WHERE id=1 AND (lease_until<NOW() OR owner=?)`，会话无关、连接池安全。leader 停续约后 TTL（30s）内备节点接管
- **下次触发计算**：`scheduler.NextTriggerTime` 支持 CRON（6 字段含秒，robfig/cron）、固定频率、固定延时

### 触发器（trigger）

- **9 种路由策略**：First/Last/Round/Random/ConsistentHash/LFU/LRU/Failover/Busyover + ShardingBroadcast，策略模式注册到 `map[int8]Router`
- **阻塞策略**：serial（串行）/ discard-later（丢弃后续）/ cover-early（覆盖之前），`BlockTracker` 跟踪每个 job 的运行中实例
- **下发**：通过 `ExecutorConn.Send`（chan）→ `ExecutorLink.Link` 的 sender goroutine → `stream.Send` 下发 `TriggerTask`

### 通信层（rpc）

- **ExecutorLink 双向流**：执行器↔admin 单一长连接复用 register/heartbeat/task_result/log_chunk/load
- **AdminService**：一元 RPC（TriggerTask/QueryTaskStatus）+ 服务端流（StreamLogs）
- **h2c**：admin rpcServer 用 `h2c.NewHandler` 支持 cleartext HTTP/2（gRPC 必需）

### 注册中心（registry）

- **纯内存**：`map[app]map[id]*ExecutorConn`，gRPC 连接建立即注册、断连即摘除
- **零外部依赖**：无 Redis、无 DB 轮询，下线感知秒级
- **负载上报**：执行器 `LoadReport` 原子写入 conn，`busyover` 路由器读取

### 持久层（store）

- **MySQL 8.0**，8 张表：users/executors/jobs/task_runs（按天分区）/scheduler_lease/audit_logs/alarm_rules
- **GORM** AutoMigrate（开发）+ `migrations/001_init.sql`（生产分区表）
- **租约锁**：`scheduler_lease` 单行 id=1，`store.Seed` 初始化

## 关键设计决策

| 决策 | 理由 |
|---|---|
| 不兼容 Java xxl-job 协议 | 摆脱 HTTP+JSON 历史包袱，用 gRPC 双向流获得长连接/实时日志/背压路由 |
| 无 Redis | MySQL 租约锁 + 内存注册中心，依赖面最小，运维成本最低 |
| gRPC over h2c | 双向流需 HTTP/2，h2c 避免 TLS 证书复杂度 |
| ConnectRPC | 一份 proto 同时服务 gRPC（执行器）与 Connect 协议（未来前端 connect-web） |
| 响应格式 `{code:0,data,msg}` | 适配 vue-admin http 拦截器约定 |

## 数据流

### 调度触发
```
scheduler.fireSecond → trigger.Trigger
  → BlockTracker.Allow（阻塞检查）
  → db.First(executor, job_group) → reg.Conns(app_name)
  → Router.Route（选目标）
  → createTaskRun（落 task_runs）
  → conn.Send <- TriggerTask（下发）
  → 更新 trigger_next_time
```

### 结果回传
```
executor 执行 handler → conn.Send <- TaskResult
  → admin ExecutorLink.Receive → handleTaskResult
  → UPDATE task_runs SET handle_code/msg/time
  → BlockTracker.OnDone（释放槽位）
  → 失败时 alarm.Notify + metrics.TriggerFailTotal
```

### 实时日志
```
executor jc.Log → conn.Send <- LogChunk
  → admin ExecutorLink.Receive → logstream.Hub.Publish(logID, content)
  → 前端 WebSocket /api/logs/:id/stream 订阅 → 推送
```

## 高可用

- **admin 集群**：多副本部署，MySQL 租约锁保证唯一 leader 调度，leader 宕机 TTL 内自动 failover
- **执行器集群**：同一 `app_name` 多实例，路由策略自动选择，故障自动转移（Failover）
- **MySQL**：主从复制，admin 连主库

## 可观测性

- **metrics**：`/metrics`（Prometheus），`xxl_job_trigger_total`/`fail_total`/`running`/`executor_online`
- **审计**：`auditMiddleware` 记录非 GET 操作到 `audit_logs`
- **告警**：任务失败触发 `alarm.Notifier`（LogNotifier/EmailNotifier，可扩展钉钉/企微）
- **日志**：zap 结构化日志（json/console）
