package db

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"http-services/db/msqldb"
)

type Migrator struct {
	Name    string
	Migrate func(*gorm.DB) error
}

func MigrateAll() error {
	database, err := msqldb.Client()
	if err != nil {
		return fmt.Errorf("init mysql client: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("get mysql sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	// 新增业务表域后，在这里按依赖顺序注册迁移：
	// return RunMigrators(database, Migrator{Name: "user", Migrate: user.Migrate})
	return RunMigrators(database)
}

func RunMigrators(database *gorm.DB, migrators ...Migrator) error {
	if database == nil {
		return errors.New("mysql client is nil")
	}

	for _, migrator := range migrators {
		if migrator.Migrate == nil {
			return fmt.Errorf("migrator %q has nil migrate function", migrator.Name)
		}

		zap.L().Info("running database migration", zap.String("module", migrator.Name))
		if err := migrator.Migrate(database); err != nil {
			zap.L().Error("database migration failed",
				zap.String("module", migrator.Name),
				zap.Error(err),
			)
			return fmt.Errorf("migrate %s: %w", migrator.Name, err)
		}
	}

	return nil
}
