// Package model 定义 GORM 持久化模型，对应 migrations/001_init.sql。
package model

import "time"

// User 用户
type User struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password   string    `gorm:"size:128;not null" json:"-"`     // bcrypt
	Role       int8      `gorm:"not null;default:0" json:"role"` // 0=普通用户 1=管理员
	Permission string    `gorm:"size:256;not null;default:''" json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// Executor 执行器分组
type Executor struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AppName     string    `gorm:"size:64;uniqueIndex;not null;column:app_name" json:"app_name"`
	Title       string    `gorm:"size:128;not null;default:''" json:"title"`
	AddressType int8      `gorm:"not null;default:0;column:address_type" json:"address_type"` // 0=自动注册 1=手动录入
	AddressList string    `gorm:"type:text;column:address_list" json:"address_list"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Executor) TableName() string { return "executors" }

// Job 任务
type Job struct {
	ID                     uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobGroup               uint64    `gorm:"not null;column:job_group;index:idx_group" json:"job_group"`
	JobDesc                string    `gorm:"size:128;not null;default:'';column:job_desc" json:"job_desc"`
	ScheduleType           int8      `gorm:"not null;default:0;column:schedule_type" json:"schedule_type"` // 0=NONE 1=CRON 2=固定频率 3=固定延时
	ScheduleConf           string    `gorm:"size:128;not null;default:'';column:schedule_conf" json:"schedule_conf"`
	ExecutorHandler        string    `gorm:"size:256;not null;default:'';column:executor_handler" json:"executor_handler"`
	ExecutorParam          string    `gorm:"size:512;not null;default:'';column:executor_param" json:"executor_param"`
	ExecutorRouteStrategy  int8      `gorm:"not null;default:0;column:executor_route_strategy" json:"executor_route_strategy"`
	ExecutorBlockStrategy  int8      `gorm:"not null;default:0;column:executor_block_strategy" json:"executor_block_strategy"`
	ExecutorTimeout        int       `gorm:"not null;default:0;column:executor_timeout" json:"executor_timeout"`
	ExecutorFailRetryCount int       `gorm:"not null;default:0;column:executor_fail_retry_count" json:"executor_fail_retry_count"`
	GlueType               int8      `gorm:"not null;default:0;column:glue_type" json:"glue_type"`
	GlueSource             string    `gorm:"type:mediumtext;column:glue_source" json:"glue_source"`
	TriggerStatus          int8      `gorm:"not null;default:0;column:trigger_status;index:idx_trigger,priority:1" json:"trigger_status"` // 0=停止 1=运行
	TriggerLastTime        int64     `gorm:"not null;default:0;column:trigger_last_time" json:"trigger_last_time"`                        // 毫秒
	TriggerNextTime        int64     `gorm:"not null;default:0;column:trigger_next_time;index:idx_trigger,priority:2" json:"trigger_next_time"`
	ChildJobID             string    `gorm:"size:256;not null;default:'';column:child_jobid" json:"child_jobid"`
	Cluster                string    `gorm:"size:64;not null;default:'';column:cluster" json:"cluster"`               // 流量泳道：指定发往的执行器 cluster，空=不限制
	CreatorID              uint64    `gorm:"not null;default:0;column:creator_id;index:idx_creator" json:"creator_id"` // 创建者 user_id；普通用户仅见自己创建的任务，admin 见全部
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (Job) TableName() string { return "jobs" }

// JobTriggerDedup 触发幂等去重：防止 leader 切换窗口对同一执行器同秒重复触发。
// (job_id, fire_second, executor_address) 唯一——冲突即代表该秒已触发，跳过。
type JobTriggerDedup struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID           uint64    `gorm:"not null;column:job_id;uniqueIndex:uk_dedup,priority:1" json:"job_id"`
	FireSecond      int64     `gorm:"not null;column:fire_second;uniqueIndex:uk_dedup,priority:2" json:"fire_second"`
	ExecutorAddress string    `gorm:"size:128;not null;default:'';column:executor_address;uniqueIndex:uk_dedup,priority:3" json:"executor_address"`
	CreatedAt       time.Time `json:"created_at"`
}

func (JobTriggerDedup) TableName() string { return "job_trigger_dedup" }

// TaskRun 任务执行记录（生产用按天分区表，见 migrations/001_init.sql）
type TaskRun struct {
	ID                    uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID                 uint64     `gorm:"not null;column:job_id;index:idx_job" json:"job_id"`
	ExecutorAddress       string     `gorm:"size:128;not null;default:'';column:executor_address" json:"executor_address"`
	ExecutorApp           string     `gorm:"size:64;not null;default:'';column:executor_app" json:"executor_app"`
	ExecutorHandler       string     `gorm:"size:256;not null;default:'';column:executor_handler" json:"executor_handler"`
	ExecutorParam         string     `gorm:"size:512;not null;default:'';column:executor_param" json:"executor_param"`
	ExecutorShardingParam string     `gorm:"size:32;not null;default:'';column:executor_sharding_param" json:"executor_sharding_param"`
	TriggerTime           time.Time  `gorm:"not null;column:trigger_time" json:"trigger_time"`
	TriggerCode           int16      `gorm:"not null;default:0;column:trigger_code" json:"trigger_code"`
	TriggerMsg            string     `gorm:"type:text;column:trigger_msg" json:"trigger_msg"`
	HandleTime            *time.Time `gorm:"column:handle_time" json:"handle_time"`
	HandleCode            int16      `gorm:"not null;default:0;column:handle_code" json:"handle_code"`
	HandleMsg             string     `gorm:"type:text;column:handle_msg" json:"handle_msg"`
	AlarmStatus           int8       `gorm:"not null;default:0;column:alarm_status" json:"alarm_status"`
}

func (TaskRun) TableName() string { return "task_runs" }

// SchedulerLease 调度租约（leader 选举，单行 id=1）
type SchedulerLease struct {
	ID         int8      `gorm:"primaryKey" json:"id"`
	Owner      string    `gorm:"size:64;not null;default:'';column:owner" json:"owner"`
	LeaseUntil time.Time `gorm:"not null;column:lease_until" json:"lease_until"`
}

func (SchedulerLease) TableName() string { return "scheduler_lease" }

// AuditLog 审计日志
type AuditLog struct {
	ID     uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Actor  string    `gorm:"size:64;not null;default:'';index:idx_actor" json:"actor"`
	Action string    `gorm:"size:64;not null;default:''" json:"action"`
	Target string    `gorm:"size:128;not null;default:''" json:"target"`
	Detail string    `gorm:"type:text" json:"detail"`
	At     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP(3);index:idx_at" json:"at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// AlarmRule 告警规则
type AlarmRule struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID     uint64    `gorm:"not null;default:0;column:job_id;index:idx_job" json:"job_id"`
	Type      string    `gorm:"size:32;not null;default:'email'" json:"type"`
	Config    string    `gorm:"type:json;column:config" json:"config"`
	Enabled   int8      `gorm:"not null;default:1" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

func (AlarmRule) TableName() string { return "alarm_rules" }

// XxlJobLogGlue GLUE 代码版本历史（最多保留 30 版本，支持回溯）
type XxlJobLogGlue struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID      uint64    `gorm:"not null;column:job_id;index" json:"job_id"`
	GlueType   int8      `gorm:"not null;default:0;column:glue_type" json:"glue_type"`
	GlueSource string    `gorm:"type:mediumtext;column:glue_source" json:"glue_source"`
	GlueRemark string    `gorm:"size:128;not null;default:'';column:glue_remark" json:"glue_remark"`
	Version    int       `gorm:"not null;default:0;column:version" json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

func (XxlJobLogGlue) TableName() string { return "xxl_job_logglue" }
