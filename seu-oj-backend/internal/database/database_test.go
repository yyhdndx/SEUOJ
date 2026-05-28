package database

import (
	"testing"

	"seu-oj-backend/internal/config"
)

func TestDatabaseDSNContract(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "db",
		Port:     "3306",
		User:     "user",
		Password: "pass",
		Name:     "seu_oj",
	}
	if got := cfg.DSN(); got != "user:pass@tcp(db:3306)/seu_oj?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("unexpected dsn %q", got)
	}
}
