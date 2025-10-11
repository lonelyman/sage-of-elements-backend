package apppostgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"sage-of-elements-backend/pkg/appconfig"
	"sage-of-elements-backend/pkg/applogger"
)

// gormLoggerAdapter คือ "หัวแปลงปลั๊ก" ที่ทำให้ Logger ของเราคุยกับ GORM ได้
type gormLoggerAdapter struct {
	appLogger applogger.Logger
}

// --- Implement gormlogger.Interface ---
func (l *gormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface { return l }
func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	l.appLogger.Info(fmt.Sprintf(msg, data...))
}
func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.appLogger.Warn(fmt.Sprintf(msg, data...))
}
func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	l.appLogger.Error(fmt.Sprintf(msg, data...), nil)
}
func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// ในอนาคตเราสามารถเพิ่ม Logic การ log SQL query ที่นี่ได้
}

// NewConnection คือฟังก์ชันที่ถูกอัปเกรดให้รับ Logger เข้ามา
func NewConnection(cfg appconfig.PostgresConfig, appLogger applogger.Logger) (*gorm.DB, error) {
	dsn := cfg.BuildDSN()

	newLogger := &gormLoggerAdapter{appLogger: appLogger}

	gormConfig := &gorm.Config{
		Logger: newLogger.LogMode(logger.Info), // ตั้งค่าให้ GORM ใช้ Logger ใหม่
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// ตั้งค่า Connection Pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// เราไม่ต้อง log ที่นี่แล้ว เพราะ GORM จะ log ให้เองผ่าน Adapter
	// appLogger.Info("Successfully connected to PostgreSQL", "host", cfg.Host, "dbName", cfg.DBName)

	return db, nil
}
