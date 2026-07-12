# 示例数据与场景 Demo

本目录提供各场景的 GLUE 脚本，方便快速体验调度中心的在线编程能力。

## 快速开始

推荐使用 docker compose 一键拉起 MySQL + 调度中心 + 通用执行器（内置 handler），无需手动导入数据：

```bash
JWT_SECRET=your-strong-secret docker compose -f deploy/docker-compose.yml up
```

浏览器访问 `http://localhost:8080`（admin/123456），执行器自动注册（app=default）。
完整部署细节见 [镜像部署](../../docs/guide/docker-images.md)。

如需内置示例任务（CRON/固定频率/父子任务等），启动 admin 时加 `--demo` 即可自动 seed：

```bash
cd backend && JWT_SECRET=test-secret-0123456789abcdef \
  DB_DSN="root:root@tcp(127.0.0.1:3306)/xxl_job?charset=utf8mb4&parseTime=True&loc=Local" \
  go run ./cmd/server --demo
```

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
