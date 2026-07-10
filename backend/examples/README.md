# 示例数据与场景 Demo

本目录提供可一键导入的示例数据与各场景的 GLUE 脚本，方便快速体验调度中心的全部能力。

## 快速开始

```bash
# 1. 启动调度中心（已自动 seed 管理员 admin/123456）
cd deploy && docker compose up -d

# 2. 导入示例数据
mysql -h127.0.0.1 -uroot -proot xxl_job < backend/examples/seed-demo.sql

# 3. 启动示例执行器（注册 demoJob / slowJob，连接 demo-app 分组）
cd backend && go run ./cmd/executor-sample -app demo-app

# 4. 浏览器登录 http://localhost:5173（admin/123456）
#    「任务管理」→ 选任务 → 点「启动」→「调度日志」观察
```

## 示例任务场景（seed-demo.sql）

| 场景 | schedule_type | 说明 |
|---|---|---|
| CRON 每分钟 | 1=CRON | `0 */1 * * * ?`（6 段含秒位） |
| 固定频率 30s | 2=固定频率 | 每 30 秒一次 |
| 固定延时 10s | 3=固定延时 | 上次结束后延时 10s 再触发 |
| 阻塞·丢弃 | 2 | 前次未完成时丢弃后续（阻塞策略=1） |
| 子任务链 | 2 | 父任务成功后触发子任务（`child_jobid`） |
| GLUE-Shell | 1 | 在线 Shell 脚本，无需执行器预注册 handler |

## 路由策略对照

| 值 | 策略 | 适用 |
|---|---|---|
| 1 | 第一个 | 固定主从 |
| 2 | 最后一个 | 固定主从（尾） |
| 3 | 轮询 | 均匀分摊 |
| 4 | 随机 | 简单打散 |
| 5 | 一致性 HASH | 同参数固定到同执行器 |
| 10 | 分片广播 | 大数据分片处理（shardingTotal/shardingIndex） |

## 阻塞策略

| 值 | 策略 | 行为 |
|---|---|---|
| 0 | 串行 | 排队依次执行 |
| 1 | 丢弃后续 | 前次未完成则丢弃新触发 |
| 2 | 覆盖之前 | 终止前次，执行新触发 |

## GLUE 在线代码示例（glue/）

| 文件 | GLUE 类型 | 说明 |
|---|---|---|
| `glue/shell.sh` | Shell | 系统巡检脚本 |
| `glue/python.py` | Python | 数据处理示例 |
| `glue/go.go` | Go（yaegi） | 强类型在线代码 |

> GLUE 默认关闭（`executor.enable_script=false`）。需在执行器侧开启脚本执行能力后再用，安全考量见仓库根 `SECURITY.md`。

## 本次新特性场景

- **用户隔离**：普通用户登录只见/管自己创建的任务；admin 见全部。新建用户在「用户管理」。
- **调度热生效**：运行中的任务改 CRON/频率后最多 5s 生效，无需停止再启动。
- **告警**：在 `backend/configs/config.yaml` 的 `alarm` 段填 webhook/SMTP，任务失败即推送（钉钉/企微/飞书/邮件）。
