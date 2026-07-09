// Package store 持久化抽象与 MySQL 实现。
package store

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xxl-job-admin-go/internal/conf"
	"xxl-job-admin-go/internal/model"
)

// Store 持久化接口（SOLID-D：依赖抽象而非具体实现）。
type Store interface {
	DB() *gorm.DB
	AutoMigrate() error
	Seed() error
}

type mysqlStore struct {
	db *gorm.DB
}

// New 创建 MySQL store 并配置连接池。
func New(cfg conf.DBConfig) (Store, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)
	return &mysqlStore{db: db}, nil
}

func (s *mysqlStore) DB() *gorm.DB { return s.db }

// AutoMigrate 自动建表（开发用，task_runs 非分区；生产用 migrations/001_init.sql）。
func (s *mysqlStore) AutoMigrate() error {
	return s.db.AutoMigrate(
		&model.User{},
		&model.Executor{},
		&model.Job{},
		&model.TaskRun{},
		&model.SchedulerLease{},
		&model.AuditLog{},
		&model.AlarmRule{},
		&model.XxlJobLogGlue{},
	)
}

// Seed 种子管理员 admin/123456（仅首次）+ 初始化调度租约行（id=1）。
func (s *mysqlStore) Seed() error {
	// 初始化调度租约行（locker.TryAcquire 依赖 id=1 行存在）
	if err := s.db.Where("id = ?", 1).
		Attrs(model.SchedulerLease{Owner: "", LeaseUntil: time.Unix(0, 0).UTC()}).
		FirstOrCreate(&model.SchedulerLease{}).Error; err != nil {
		return fmt.Errorf("init lease: %w", err)
	}

	var cnt int64
	if err := s.db.Model(&model.User{}).Where("username = ?", "admin").Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}
	return s.db.Create(&model.User{
		Username: "admin",
		Password: string(hash),
		Role:     1, // 管理员
	}).Error
}
