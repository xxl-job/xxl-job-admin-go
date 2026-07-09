# 安全策略

## 报告漏洞

如果你发现安全漏洞，请**不要**公开提 Issue。

请私密报告至：维护者邮箱（见仓库 owner 资料），或通过 GitHub Security Advisory 提交。

报告请包含：

- 漏洞描述与影响范围
- 复现步骤（最小化）
- 受影响版本
- 建议的修复方向

我们会在收到报告后 72 小时内响应，并在修复后致谢报告者。

## 支持版本

| 版本 | 安全支持 |
|---|---|
| 0.1.x | ✅ |
| < 0.1 | ❌ |

## 安全建议（生产部署）

- `JWT_SECRET` 使用随机 64 字符以上，通过环境变量注入，勿入库
- `xxl.job.accessToken` 调度中心与执行器之间启用访问令牌
- MySQL 账户最小权限，仅授予 `xxl_job` 库
- 生产环境关闭 `/metrics` 公网访问或加基本认证
- K8s 部署用 NetworkPolicy 限制执行器与 admin 间流量

## GLUE / 脚本任务（重要）

GLUE Go（yaegi 解释器）与 Shell/Python 脚本任务可执行**任意代码**，存在命令注入与沙箱逃逸风险。执行器 SDK 默认**禁用**脚本/GLUE 任务：

- 执行器需显式调用 `executor.WithScriptTask(true)` 才会执行 `glue_type=1/2/3` 的任务；否则返回 500 拒绝。
- 仅在受控、隔离的执行器环境开启；开启后不得接受不可信来源的 GLUE/脚本任务（配合 RBAC 限制 `updateJob`）。
- yaegi 暴露完整 Go 标准库（含 `os`/`net`），非安全沙箱；脚本经 `sh -c` / `python3 -c` 执行，等同 shell 注入面。

## 已知限制

- **refresh token / 登出**：当前 refresh token 复用 access token、登出不吊销已签发 token（无服务端 session 存储）。token 泄露后在 TTL（默认 24h）内无法吊销。生产高安全场景需自行接入 session 存储实现吊销，列入后续迭代。
- **WebSocket 日志**：默认同源校验（防 CSWSH）；跨域部署需通过 `WS_ALLOWED_ORIGINS` 环境变量显式放行。
