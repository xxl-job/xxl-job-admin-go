-- XXL-JOB-Admin-Go 示例数据：执行器分组 + 多场景任务
-- 覆盖：CRON / 固定频率 / 固定延时 / 阻塞策略 / 子任务链 / GLUE-Shell / 路由策略
--
-- 用法（开发库已建表后执行）：
--   mysql -uroot -p xxl_job < backend/examples/seed-demo.sql
-- 前置：
--   1) 调度中心已启动（自动 seed 管理员 admin/123456）
--   2) 启动示例执行器连接 demo-app 分组：
--      cd backend && go run ./cmd/executor-sample -app demo-app
-- 导入后到前端「任务管理」点「启动」即可观察调度（默认全部停止状态）。

-- 示例执行器分组（与 executor-sample 的 app_name 一致）
INSERT INTO executors (app_name, title, address_type) VALUES ('demo-app', '示例执行器', 0)
  ON DUPLICATE KEY UPDATE title = VALUES(title);
SET @gid = (SELECT id FROM executors WHERE app_name = 'demo-app');

-- 1. CRON 任务：每分钟跑 demoJob（路由=第一个，阻塞=串行）
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id)
VALUES (@gid, '示例·CRON每分钟', 1, '0 */1 * * * ?', 'demoJob', 'hello-cron', 1, 0, 1);

-- 2. 固定频率：每 30 秒（路由=轮询）
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id)
VALUES (@gid, '示例·固定频率30s', 2, '30', 'demoJob', 'hello-fixed', 3, 0, 1);

-- 3. 固定延时：slowJob，延时 10s（超时 30s）
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, executor_timeout, creator_id)
VALUES (@gid, '示例·固定延时10s', 3, '10', 'slowJob', 'slow-demo', 1, 0, 30, 1);

-- 4. 阻塞策略·丢弃后续：每 5s 触发，前次未完成则丢弃（用 slowJob 制造堆积）
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id)
VALUES (@gid, '示例·阻塞丢弃', 2, '5', 'slowJob', 'block-drop', 1, 1, 1);

-- 5. 子任务链：子任务（先建），父任务成功后触发它
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, child_jobid, creator_id)
VALUES (@gid, '示例·子任务(子)', 2, '60', 'demoJob', 'child', 1, 0, '', 1);
SET @child = LAST_INSERT_ID();
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, child_jobid, creator_id)
VALUES (@gid, '示例·子任务(父)', 2, '60', 'demoJob', 'parent', 1, 0, @child, 1);

-- 6. GLUE-Shell 在线脚本：每 2 分钟执行 Shell（glue_type=2）
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, glue_type, glue_source, executor_route_strategy, executor_block_strategy, creator_id)
VALUES (@gid, '示例·GLUE-Shell', 1, '0 */2 * * * ?', 'glueShell', 2,
  '#!/bin/bash\necho "GLUE Shell 执行: $(date)"\necho "param=$1"\nexit 0',
  1, 0, 1);

-- 路由策略对照（均停止状态，供查阅）：随机(4) / 一致性HASH(5) / 分片广播(10)
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id) VALUES (@gid,'示例·路由随机',2,'60','demoJob','route-random',4,0,1);
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id) VALUES (@gid,'示例·路由一致性HASH',2,'60','demoJob','route-hash',5,0,1);
INSERT INTO jobs (job_group, job_desc, schedule_type, schedule_conf, executor_handler, executor_param, executor_route_strategy, executor_block_strategy, creator_id) VALUES (@gid,'示例·分片广播',2,'60','demoJob','route-shard',10,0,1);

-- 导入结果（执行后回显）
SELECT id, job_desc, schedule_type, schedule_conf, executor_handler, executor_route_strategy, executor_block_strategy
FROM jobs WHERE job_group = @gid ORDER BY id;
