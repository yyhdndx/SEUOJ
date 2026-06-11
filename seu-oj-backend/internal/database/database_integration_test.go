package database

import (
	"testing"

	"seu-oj-backend/internal/config"
)

func TestNewWithConfiguredMySQL(t *testing.T) {
	cfg := config.Load()
	db, err := New(cfg.Database)
	if err != nil {
		t.Skipf("mysql unavailable for integration test: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
}
