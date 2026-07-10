# 运维与可观测

面向 **SRE / 平台运维**：配置参数、监控指标、日志、租约锁高可用、健康检查与故障排查。

## 配置与启动

调度中心无命令行 flag，通过 `configs/config.yaml` 加载，环境变量以 `SECTION_KEY`（点转下划线、全大写）覆盖。配置文件路径用 `CONFIG_PATH` 指定，默认 `configs/config.yaml`。

### 关键配置项

| 配置项 | 环境变量 | 默认 | 说明 |
|---|---|---|---|
| `server.http_addr` | `SERVER_HTTP_ADDR` | `:8080` | REST API + `/metrics` + `/healthz` |
| `server.grpc_addr` | `SERVER_GRPC_ADDR` | `:9090` | gRPC，执行器 `ExecutorLink` |
| `db.dsn` | `DB_DSN` | — | MySQL DSN（GORM） |
| `jwt.secret` | `JWT_SECRET` | — | **启动强制校验**：非空 / ≥16 字节 / 非占位符，否则拒绝启动 |
| `jwt.ttl` | `JWT_TTL` | `24h` | Access Token 有效期 |
| `lock.lease_ttl` | `LOCK_LEASE_TTL` | `30s` | leader 租约有效期 |
| `lock.renew_interval` | `LOCK_RENEW_INTERVAL` | `10s` | leader 续约间隔 |
| `log.level` | `LOG_LEVEL` | `info` | debug/info/warn/error |
| `log.format` | `LOG_FORMAT` | `json` | json / console |
| `node_id` | `NODE_ID` | 自动 `hostname-pid` | 节点标识（租约 owner） |
| `access_token` | `ACCESS_TOKEN` | 空 | 调度中心↔执行器认证令牌，空则不校验 |

> Docker Compose 中 `JWT_SECRET` 写作 `${JWT_SECRET:?...}`，未设置时 `docker compose up` 直接拒绝启动，避免误用弱密钥。启动前执行 `export JWT_SECRET="$(openssl rand -hex 32)"`。

## 监控指标（Prometheus）

`/metrics` 暴露在 HTTP 端口（默认 `:8080`），四个核心指标：

| 指标 | 类型 | 标签 | 含义 |
|---|---|---|---|
| `xxl_job_trigger_total` | Counter | `handler` | 任务触发总数 |
| `xxl_job_trigger_fail_total` | Counter | `handler` | 触发失败总数 |
| `xxl_job_running` | Gauge | — | 当前运行中任务数 |
| `xxl_job_executor_online` | Gauge | `app` | 在线执行器数（按 app） |

### 常用 PromQL

```txt
# 各 handler 触发速率
sum by (handler) (rate(xxl_job_trigger_total[1m]))

# 各 handler 触发失败率
sum by (handler) (rate(xxl_job_trigger_fail_total[5m]))
  / sum by (handler) (rate(xxl_job_trigger_total[5m]))

# 当前运行中任务数
xxl_job_running

# 各执行器分组在线数
xxl_job_executor_online
```

建议告警：某 handler 失败率超阈值持续 N 分钟；某 `app` 的 `executor_online` 跌至 0（执行器全部离线，任务将无法调度）。

## 日志

zap 结构化日志，默认 `json` / `info`，每条含 `node_id`（节点标识），便于多副本环境区分来源。开发调试设 `LOG_FORMAT=console LOG_LEVEL=debug`。

容器查看：

```bash
docker compose logs -f admin
# 或 Kubernetes
kubectl logs -f deploy/xxl-job-admin -c admin
```

## 租约锁与高可用（HA）

调度中心用 **MySQL 单行租约表**选主，无需 Redis：

- 租约表 `scheduler_lease`（`id=1` 唯一行），抢锁/续约用原子 SQL：

  ```sql
  UPDATE scheduler_lease
  SET owner=?, lease_until=DATE_ADD(NOW(3), INTERVAL ? SECOND)
  WHERE id=1 AND (lease_until < NOW(3) OR owner=?)
  ```

- 有效期 `lease_ttl=30s`，每 `renew_interval=10s` 续约一次。

### 多副本行为

- **leader**：持租约的副本，独占运行任务调度器与分区维护器。
- **非 leader**：提供完整 REST/gRPC 服务、接收执行器连接，但不触发调度，待机等待接管。

### 故障转移

leader 停止续约 → 最多 30s 后租约过期 → 备副本抢锁成功 → 新 leader 启动调度。原 leader 若恢复，发现租约已被占则降级为备。

> 故障转移窗口最长为 `lease_ttl`（30s），期间可能重复触发或漏触发，需结合任务幂等性设计。生产多副本建议 `replicas=2`（k8s 模板默认）+ 跨可用区部署。

## 健康检查

`/healthz` 返回 HTTP 200 `{"status":"ok"}`，用于容器探针：

```yaml
livenessProbe:
  httpGet: { path: /healthz, port: 8080 }
  initialDelaySeconds: 10
  periodSeconds: 15
readinessProbe:
  httpGet: { path: /healthz, port: 8080 }
  initialDelaySeconds: 5
  periodSeconds: 10
```

> `/healthz` 仅反映进程存活，不含 DB 连接与 leader 状态。调度可用性以 metrics + 日志为准。

## 数据分区维护

调度日志按日分区，leader 每 24h 维护一次：保留近 30 天、预建未来 7 天分区。无需人工干预。

## 故障排查

### 启动失败：JWT_SECRET 校验

日志含「jwt.secret 不能为空 / 强度不足 / 不能使用默认占位符」。设置强密钥：`export JWT_SECRET="$(openssl rand -hex 32)"`。

### 启动失败：连接 MySQL

`DB_DSN` 错误或 MySQL 未就绪。Compose 已配 `depends_on: service_healthy`；二进制部署需确认 MySQL 可达且已执行 `migrations/`。

### 任务不触发

1. 执行器是否在线：`xxl_job_executor_online{app="<分组>"}` > 0？为 0 说明执行器未注册或离线。
2. 本节点是否 leader：日志中触发节点的 `node_id` 与目标节点比对；非 leader 不调度属正常。
3. CRON 表达式：6 段含秒位（`秒 分 时 日 月 周`），如 `*/5 * * * * ?`，而非 5 段。

### 触发成功但执行失败

查调度日志 `handle_msg`；多为执行器侧 handler 未注册 / panic / 超时。实时日志走 WebSocket（「查看日志」）。

### metrics 抓不到

确认 Prometheus 抓取目标指向 HTTP 端口（默认 `8080`）的 `/metrics`：`curl http://<pod-ip>:8080/metrics` 应返回文本。

## 扩缩容与接管

- **扩容**：直接加副本，新副本竞争租约，未抢到则待机，对在途调度无影响。
- **缩容**：缩掉 leader 时等待 ≤30s 租约过期后自动转移；建议滚动缩容。
- **滚动更新**：逐个替换，确保任意时刻至少一个副本持租约，避免调度空窗。

---

相关：[部署指南](./design/deployment) · [架构设计](./design/architecture) · [执行器开发](./design/executor-dev)
