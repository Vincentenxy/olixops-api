// Package database 封装 GORM 数据库连接与事务工具。
package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"olixops/internal/config"
)

// New 创建并初始化 GORM 数据库实例。
func New(cfg config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	if !strings.EqualFold(cfg.Driver, "postgres") {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	gormCfg := &gorm.Config{
		Logger:                                   newGormLogger(log, cfg.LogLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("acquire sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func newGormLogger(zlog *zap.Logger, level string) gormlogger.Interface {
	gl := gormlogger.New(
		zapPrintf{logger: zlog},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  parseGormLevel(level),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	return gl
}

func parseGormLevel(level string) gormlogger.LogLevel {
	switch strings.ToLower(level) {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn", "warning":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

// zapPrintf 适配 GORM logger 的 Printf 接口到 Zap。
type zapPrintf struct {
	logger *zap.Logger
}

func (z zapPrintf) Printf(format string, args ...any) {
	z.logger.Sugar().Infof(format, args...)
}
