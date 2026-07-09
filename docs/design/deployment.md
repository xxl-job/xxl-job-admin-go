# 部署指南

## 前置条件

- MySQL 8.0+
- （可选）Docker / Kubernetes

## 一、Docker Compose（推荐起步）

```bash
docker-compose up -d          # 启动 MySQL + admin
go run ./cmd/executor-sample  # 示例执行器
cd web && pnpm dev            # 前端
```

访问 `http://localhost:5173`，登录 `admin/123456`。

`docker-compose.yml` 已配置 MySQL 健康检查，admin 依赖 MySQL healthy 后启动。环境变量覆盖配置（`DB_DSN`、`JWT_SECRET` 等）。

## 二、二进制部署

```bash
make build                    # 生成 bin/xxl-job-admin-go
./bin/xxl-job-admin-go        # 前台运行（CONFIG_PATH=configs/config.yaml）
```

生产配置用 `configs/config.prod.yaml`，敏感项通过环境变量注入：

```bash
DB_DSN="user:pwd@tcp(mysql:3306)/xxl_job?..." \
JWT_SECRET="random-64-chars" \
./bin/xxl-job-admin-go
```

## 三、Kubernetes

`k8s/admin.yaml` 提供 Deployment + Service：

```bash
kubectl apply -f k8s/admin.yaml
```

特性：
- `replicas: 2` 多副本 HA（MySQL 租约锁保证唯一 leader 调度）
- `livenessProbe`/`readinessProbe` 走 `/healthz`
- 敏感配置通过 Secret 注入

需先创建 Secret：
```bash
kubectl create secret generic xxl-job-secret \
  --from-literal=dsn='user:pwd@tcp(mysql:3306)/xxl_job?...' \
  --from-literal=jwt='random-64-chars'
```

## 四、数据库初始化

开发环境 admin 启动时 `AutoMigrate` 自动建表（非分区）。

生产环境用 `migrations/001_init.sql`（含 `task_runs` 按天分区）：

```bash
mysql -h <host> -u root -p < schemas/xxl_job < migrations/001_init.sql
```

分区自动维护：leader 节点每 24 小时执行 `ManagePartitions`（REORGANIZE p_max 预建未来 7 天 + DROP 30 天前老分区）。

## 五、执行器部署

Go 应用集成执行器 SDK：

```go
ex := executor.New("http://admin:9090", "my-app", logger)
ex.Register("handlerName", func(ctx, jc, param) (int, string) {
    jc.Log("执行中...")
    return 200, "ok"
})
ex.Run(ctx)
```

详见 [执行器开发指南](./executor-dev.md)。

## 六、生产清单

- [ ] `JWT_SECRET` 使用随机 64+ 字符
- [ ] MySQL 账户最小权限（仅 `xxl_job` 库）
- [ ] `/metrics` 端点加访问控制（Ingress basic auth 或 NetworkPolicy）
- [ ] admin 多副本部署（HA）
- [ ] 执行器集群部署（同 `app_name` 多实例）
- [ ] 配置日志收集（json 格式 → ELK/Loki）
- [ ] 配置 Prometheus 抓取 `/metrics`
- [ ] `task_runs` 分区表定期维护（内置自动）
- [ ] 数据库定期备份
