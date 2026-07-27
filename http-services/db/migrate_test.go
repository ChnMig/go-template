package db

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func TestRunMigrators(t *testing.T) {
	database := &gorm.DB{}
	var order []string

	err := RunMigrators(
		database,
		Migrator{Name: "first", Migrate: func(db *gorm.DB) error {
			if db != database {
				t.Fatalf("migrator received unexpected db")
			}
			order = append(order, "first")
			return nil
		}},
		Migrator{Name: "second", Migrate: func(db *gorm.DB) error {
			order = append(order, "second")
			return nil
		}},
	)
	if err != nil {
		t.Fatalf("RunMigrators() error = %v", err)
	}

	want := []string{"first", "second"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("migration order = %v, want %v", order, want)
	}
}

func TestRunMigratorsWithNilDB(t *testing.T) {
	if err := RunMigrators(nil); err == nil {
		t.Fatal("RunMigrators(nil) error = nil, want error")
	}
}

func TestRunMigratorsWithNilMigrator(t *testing.T) {
	err := RunMigrators(&gorm.DB{}, Migrator{Name: "missing"})
	if err == nil {
		t.Fatal("RunMigrators() error = nil, want error")
	}
}

func TestRunMigratorsWrapsMigratorError(t *testing.T) {
	wantErr := errors.New("boom")
	err := RunMigrators(&gorm.DB{}, Migrator{Name: "broken", Migrate: func(db *gorm.DB) error {
		return wantErr
	}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunMigrators() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestMigrateAllRequiresLifecycleOwnedDatabase(t *testing.T) {
	if err := MigrateAll(context.Background(), nil); !errors.Is(err, gorm.ErrInvalidData) {
		t.Fatalf("MigrateAll() error = %v, want %v", err, gorm.ErrInvalidData)
	}
}
