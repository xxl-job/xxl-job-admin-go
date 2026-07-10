-- XXL-JOB-Admin-Go 初始 schema（MySQL 8.0）
-- 生产初始化脚本（含 task_runs 按天分区）。开发环境可用 GORM AutoMigrate 建非分区表。

-- 用户
CREATE TABLE IF NOT EXISTS users (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username   VARCHAR(64)  NOT NULL,
  password   VARCHAR(128) NOT NULL,              -- bcrypt
  role       TINYINT      NOT NULL DEFAULT 0,    -- 0=普通用户 1=管理员
  permission VARCHAR(256) NOT NULL DEFAULT '',   -- 执行器权限（逗号分隔 app_name）
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 执行器分组
CREATE TABLE IF NOT EXISTS executors (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  app_name     VARCHAR(64)  NOT NULL,
  title        VARCHAR(128) NOT NULL DEFAULT '',
  address_type TINYINT      NOT NULL DEFAULT 0,  -- 0=自动注册 1=手动录入
  address_list TEXT,
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_app (app_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 任务
CREATE TABLE IF NOT EXISTS jobs (
  id                          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_group                   BIGINT UNSIGNED NOT NULL,            -- 关联 executors.id
  job_desc                    VARCHAR(128) NOT NULL DEFAULT '',
  schedule_type               TINYINT      NOT NULL DEFAULT 0,     -- 0=NONE 1=CRON 2=固定频率 3=固定延时
  schedule_conf               VARCHAR(128) NOT NULL DEFAULT '',
  executor_handler            VARCHAR(256) NOT NULL DEFAULT '',
  executor_param              VARCHAR(512) NOT NULL DEFAULT '',
  executor_route_strategy     TINYINT      NOT NULL DEFAULT 0,     -- 路由策略
  executor_block_strategy     TINYINT      NOT NULL DEFAULT 0,     -- 阻塞策略
  executor_timeout            INT          NOT NULL DEFAULT 0,     -- 秒
  executor_fail_retry_count   INT          NOT NULL DEFAULT 0,
  glue_type                   TINYINT      NOT NULL DEFAULT 0,
  glue_source                 MEDIUMTEXT,
  trigger_status              TINYINT      NOT NULL DEFAULT 0,     -- 0=停止 1=运行
  trigger_last_time           BIGINT       NOT NULL DEFAULT 0,
  trigger_next_time           BIGINT       NOT NULL DEFAULT 0,
  child_jobid                 VARCHAR(256) NOT NULL DEFAULT '',
  cluster                     VARCHAR(64)  NOT NULL DEFAULT '',    -- 流量泳道：目标执行器 cluster，空=不限制
  creator_id                  BIGINT UNSIGNED NOT NULL DEFAULT 0,  -- 创建者 user_id；普通用户仅见自己创建的任务，admin 见全部
  created_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_trigger (trigger_status, trigger_next_time),
  KEY idx_group (job_group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 任务执行记录（按天分区，定期加新分区、删老分区）
CREATE TABLE IF NOT EXISTS task_runs (
  id                     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id                 BIGINT UNSIGNED NOT NULL,
  executor_address       VARCHAR(128) NOT NULL DEFAULT '',
  executor_handler       VARCHAR(256) NOT NULL DEFAULT '',
  executor_param         VARCHAR(512) NOT NULL DEFAULT '',
  executor_sharding_param VARCHAR(32) NOT NULL DEFAULT '',
  trigger_time           DATETIME(3)  NOT NULL,
  trigger_code           SMALLINT     NOT NULL DEFAULT 0,
  trigger_msg            TEXT,
  handle_time            DATETIME(3)  NULL,
  handle_code            SMALLINT     NOT NULL DEFAULT 0,
  handle_msg             TEXT,
  alarm_status           TINYINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id, trigger_time),
  KEY idx_job (job_id),
  KEY idx_trigger (trigger_time, trigger_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
PARTITION BY RANGE (TO_DAYS(trigger_time)) (
  PARTITION p20260707 VALUES LESS THAN (TO_DAYS('2026-07-08')),
  PARTITION p_max    VALUES LESS THAN MAXVALUE
);

-- 调度租约（leader 选举）
CREATE TABLE IF NOT EXISTS scheduler_lease (
  id          TINYINT     NOT NULL DEFAULT 1,
  owner       VARCHAR(64) NOT NULL DEFAULT '',
  lease_until DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT chk_single CHECK (id = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO scheduler_lease (id, owner, lease_until) VALUES (1, '', '1970-01-01 00:00:00');

-- 审计日志
CREATE TABLE IF NOT EXISTS audit_logs (
  id     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  actor  VARCHAR(64)  NOT NULL DEFAULT '',
  action VARCHAR(64)  NOT NULL DEFAULT '',
  target VARCHAR(128) NOT NULL DEFAULT '',
  detail TEXT,
  at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_actor (actor),
  KEY idx_at (at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 告警规则
CREATE TABLE IF NOT EXISTS alarm_rules (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id     BIGINT UNSIGNED NOT NULL DEFAULT 0,  -- 0=全局
  type       VARCHAR(32) NOT NULL DEFAULT 'email',
  config     JSON,
  enabled    TINYINT     NOT NULL DEFAULT 1,
  created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_job (job_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
