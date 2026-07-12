// Package conf 负责配置加载与环境变量覆盖。
package conf

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 顶层配置结构。
type Config struct {
	Server        ServerConfig `mapstructure:"server"`
	DB            DBConfig     `mapstructure:"db"`
	JWT           JWTConfig    `mapstructure:"jwt"`
	Lock          LockConfig   `mapstructure:"lock"`
	Log           LogConfig    `mapstructure:"log"`
	Alarm         AlarmConfig  `mapstructure:"alarm"`
	NodeID        string       `mapstructure:"node_id"`
	AccessToken   string       `mapstructure:"access_token"`    // 调度中心↔执行器通信认证
	LogRetainDays int          `mapstructure:"log_retain_days"` // task_runs 保留天数，0=不清理（缺省 30）
}

// ServerConfig HTTP/gRPC 监听地址。
type ServerConfig struct {
	HTTPAddr string `mapstructure:"http_addr"`
	GRPCAddr string `mapstructure:"grpc_addr"`
}

// DBConfig 数据库连接。
type DBConfig struct {
	Driver       string `mapstructure:"driver"`
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// JWTConfig 认证令牌。
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	TTL    time.Duration `mapstructure:"ttl"`
}

// LockConfig 分布式租约锁。
type LockConfig struct {
	LeaseTTL      time.Duration `mapstructure:"lease_ttl"`
	RenewInterval time.Duration `mapstructure:"renew_interval"`
}

// LogConfig 日志。
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// AlarmConfig 告警通道：留空则不启用该渠道（默认仅日志告警）。
type AlarmConfig struct {
	Dingtalk string     `mapstructure:"dingtalk"` // 钉钉机器人 webhook URL
	Wechat   string     `mapstructure:"wechat"`   // 企业微信机器人 webhook URL
	Feishu   string     `mapstructure:"feishu"`   // 飞书机器人 webhook URL
	Email    EmailAlarm `mapstructure:"email"`
}

// EmailAlarm 邮件告警（SMTP）。
type EmailAlarm struct {
	Host     string `mapstructure:"host"` // SMTP 主机，如 smtp.qq.com
	Port     int    `mapstructure:"port"` // 465(SSL)/587
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	To       string `mapstructure:"to"` // 逗号分隔多个收件人
}

// Load 从 path 读取配置，环境变量以 `SECTION_KEY`（点转下划线、全大写）覆盖。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	// log_retain_days 哨兵：未设置→默认 30；显式配置 0→真 0（关闭清理）
	if !v.IsSet("log_retain_days") {
		c.LogRetainDays = 30
	}
	if c.NodeID == "" {
		hostname, _ := os.Hostname()
		c.NodeID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}
	if err := validateJWTSecret(c.JWT.Secret); err != nil {
		return nil, err
	}
	return &c, nil
}

// weakSecrets 已知的默认占位符密钥，禁止用于任何环境。
var weakSecrets = map[string]struct{}{
	"change-me-in-production": {},
	"CHANGE_ME":               {},
}

// validateJWTSecret 校验 JWT 密钥强度：非空、长度 ≥16、非默认占位符。
// 启动期强校验，防止误用示例密钥导致 token 可被伪造。
func validateJWTSecret(secret string) error {
	switch {
	case secret == "":
		return fmt.Errorf("jwt.secret 不能为空，请设置强密钥或用 JWT_SECRET 环境变量覆盖")
	case len(secret) < 16:
		return fmt.Errorf("jwt.secret 强度不足：长度需 ≥16 字节，当前 %d", len(secret))
	}
	if _, weak := weakSecrets[secret]; weak {
		return fmt.Errorf("jwt.secret 不能使用默认占位符（change-me-in-production/CHANGE_ME），请设置强密钥")
	}
	return nil
}
