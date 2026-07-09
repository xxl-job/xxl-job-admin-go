# API 参考

所有 REST 响应格式：`{"code":0,"data":<T>,"msg":"ok"}`（code 0 成功，非 0 失败）。需认证的接口在 Header 带 `Authorization: Bearer <accessToken>`。

## 认证

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/auth/sessions` | 登录，body `{username,password}` → `{accessToken,refreshToken,expiresIn}` |
| POST | `/api/auth/tokens/refresh` | 刷新 token，body `{refreshToken}` |
| DELETE | `/api/auth/sessions` | 登出 |
| GET | `/api/auth/users/me` | 当前用户信息 `{id,username,nickname,roles[],permissions[]}` |

## 任务管理

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/jobs?group=&page=&size=` | 任务列表 |
| GET | `/api/jobs/:id` | 任务详情 |
| POST | `/api/jobs` | 创建任务 |
| PUT | `/api/jobs/:id` | 更新任务 |
| DELETE | `/api/jobs/:id` | 删除任务 |
| POST | `/api/jobs/:id/start` | 启动（计算 trigger_next_time） |
| POST | `/api/jobs/:id/stop` | 停止 |
| POST | `/api/jobs/:id/trigger` | 手动触发一次 |
| GET | `/api/jobs/:id/glue-versions` | GLUE 代码版本历史（最近 30） |

任务字段：`job_group`(executor id)、`schedule_type`(1=CRON/2=固定频率/3=固定延时)、`schedule_conf`、`executor_handler`、`executor_param`、`executor_route_strategy`、`executor_block_strategy`、`executor_timeout`、`executor_fail_retry_count`。

## 执行器管理

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/executors` | 列表（含在线机器） |
| POST | `/api/executors` | 创建分组 |
| PUT | `/api/executors/:id` | 更新 |
| DELETE | `/api/executors/:id` | 删除 |

字段：`app_name`、`title`、`address_type`(0=自动注册/1=手动录入)、`address_list`。

## 日志

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/logs?job_id=&status=&page=&size=` | 调度日志列表（status=success/fail） |
| GET | `/api/logs/:id/stream?token=xxx` | WebSocket 实时日志（token 经 query，浏览器 WS 不能自定义 header） |

## 仪表盘 / 审计 / 用户

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/dashboard` | 运行报表（job_total/running/exec_total/today_*) |
| GET | `/api/audit?page=&size=` | 审计日志 |
| GET | `/api/users` | 用户列表 |
| POST | `/api/users` | 创建用户（bcrypt） |
| PUT | `/api/users/:id` | 更新（密码留空不修改） |
| DELETE | `/api/users/:id` | 删除（禁删 id=1） |

## 运维

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 健康检查 |
| GET | `/metrics` | Prometheus 指标 |

## gRPC（执行器 ↔ admin）

`ExecutorLink.Link`（双向流）：
- 执行器→admin：`ExecutorMessage`（oneof: register/heartbeat/task_result/log_chunk/load）
- admin→执行器：`AdminMessage`（oneof: register_ack/trigger/cancel/config）

`AdminService`：
- `TriggerTask(TriggerTaskRequest) → TriggerTaskResponse`
- `StreamLogs(LogRequest) → stream LogChunk`
- `QueryTaskStatus(StatusRequest) → StatusResponse`

proto 定义见 `proto/xxl/v1/`。
