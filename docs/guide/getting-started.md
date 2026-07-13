# 快速开始

5 分钟启动调度中心、登录、注册执行器并调度第一个任务。

## Docker 快速开始（推荐）

```bash
JWT_SECRET=your-strong-secret docker compose -f deploy/docker-compose.yml up
```

浏览器访问 `http://localhost:8080`（admin/123456）。详见 [镜像部署](./docker-images)。

::: tip 仅体验？
不想单独部署 executor？用 [all-in-one 2 容器 compose](./docker-images#一键体验-all-in-one-2-容器)：`JWT_SECRET=xxx docker compose -f deploy/docker-compose.allinone.yml up`，admin 内嵌通用执行器，开箱即用。
:::

---

## 1. 前置条件

- Docker + Docker Compose（推荐起步）
- 或本地 Go 1.25+ 与 MySQL 8.0+（二进制启动见[镜像部署](./docker-images)）

## 2. 启动调度中心（Docker Compose）

```bash
# 设置 JWT 密钥（≥16 字节；启动校验拒绝弱密钥与占位符）
export JWT_SECRET="$(openssl rand -hex 32)"

# 一键拉起 MySQL + 调度中心
cd deploy && docker compose up -d
```

验证就绪：

```bash
docker compose ps                   # admin 状态为 healthy
curl http://localhost:8080/healthz  # {"status":"ok"}
```

端口约定：

| 端口 | 协议 | 用途 |
|---|---|---|
| `8080` | HTTP | REST API + `/metrics` + `/healthz` |
| `9090` | gRPC | 执行器 `ExecutorLink` 连接 |

## 3. 启动前端（开发模式）

```bash
cd frontend && pnpm install && pnpm dev
```

浏览器打开 `http://localhost:5173`，dev server 已将 `/api` 代理到 `localhost:8080`。

::: tip 生产部署
生产环境前端独立构建（`pnpm build`）后由 nginx 托管，REST 指向调度中心 `8080`。详见[镜像部署](./docker-images)。
:::

## 4. 登录系统

内置管理员账号：

| 项 | 值 |
|---|---|
| 用户名 | `admin` |
| 密码 | `123456` |

::: warning
首次登录后请立即在「用户管理」修改默认密码。
:::

## 5. 注册执行器分组

任务必须调度到「执行器分组」上：

1. 进入「执行器管理」→「新建」
2. 填写 `app_name`（执行器唯一标识，如 `demo-app`）与标题
3. 地址类型选「自动注册」（执行器启动后自动上报）

执行器侧需用相同 `app_name` 经 gRPC `:9090` 接入，详见[执行器管理](./executor)。

## 6. 创建第一个任务

1. 进入「任务管理」→「新建」
2. 选择上一步的执行器分组
3. 填写：
   - **Handler**：执行器注册的处理函数名（如 `demoJob`）
   - **调度类型**：CRON / 固定频率 / 固定延时
   - **调度配置**：如 CRON `*/5 * * * * ?`（每 5 秒，6 段含秒位）
   - **参数**：传给 handler 的字符串
4. 保存后点「启动」

## 7. 查看调度结果

进入「调度日志」，可见每次触发：

- 触发时间、触发结果、被调度的机器地址
- 执行结果（成功/失败）、`handle_msg`、耗时
- 点「查看日志」走 WebSocket 实时流

---

至此完成首个任务的完整调度闭环。后续：

- [任务管理](./job) · [调度日志](./log) · [用户与权限](./user)
- [运维与可观测](../operations)（监控指标 / 租约锁 HA / 故障排查）
- [镜像部署](./docker-images)
