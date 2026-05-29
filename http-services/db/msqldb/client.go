package msqldb

import (
	"errors"
	stdlog "log"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"http-services/config"
	httplog "http-services/utils/log"
)

var (
	client   *gorm.DB
	clientMu sync.Mutex
)

var ErrMissingMysqlDSN = errors.New("database.mysql_dsn is empty")

func Client() (*gorm.DB, error) {
	if err := Init(); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("mysql client not initialized")
	}
	return client, nil
}

func GetClient() *gorm.DB {
	database, err := Client()
	if err != nil {
		zap.L().Error("mysql client not initialized", zap.Error(err))
		return nil
	}
	return database
}

func Init() error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if client != nil {
		return nil
	}
	if strings.TrimSpace(config.MysqlDSN) == "" {
		return ErrMissingMysqlDSN
	}

	zapLogger := zap.L().With(zap.String("component", "gorm"))
	gormWriter := httplog.NewZapWriter(zapLogger, zapcore.InfoLevel)
	gormLogger := logger.New(
		stdlog.New(gormWriter, "\r\n", stdlog.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	database, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       config.MysqlDSN,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger:                                   gormLogger,
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		DisableNestedTransaction:                 true,
		CreateBatchSize:                          1000,
		DisableAutomaticPing:                     true,
		QueryFields:                              true,
	})
	if err != nil {
		zap.L().Error("connect to mysql failed", zap.Error(err))
		return err
	}

	sqlDB, err := database.DB()
	if err != nil {
		zap.L().Error("get mysql sql.DB failed", zap.Error(err))
		return err
	}

	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	client = database
	return nil
}

func CloseClient() {
	clientMu.Lock()
	defer clientMu.Unlock()

	if client == nil {
		return
	}

	sqlDB, err := client.DB()
	if err != nil {
		zap.L().Warn("get mysql sql.DB before close failed", zap.Error(err))
		client = nil
		return
	}
	if err := sqlDB.Close(); err != nil {
		zap.L().Warn("close mysql client failed", zap.Error(err))
	}
	client = nil
}
