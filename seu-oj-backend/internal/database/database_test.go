package database

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"seu-oj-backend/internal/config"
)

func TestNewRedisConnects(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewRedis(config.RedisConfig{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("new redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
}

func TestNewRedisRejectsUnreachable(t *testing.T) {
	if _, err := NewRedis(config.RedisConfig{Addr: "127.0.0.1:1"}); err == nil {
		t.Fatal("expected redis ping failure")
	}
}

func TestDatabaseConfigDSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		User: "root", Password: "secret", Host: "localhost", Port: "3306", Name: "seuoj",
	}
	dsn := cfg.DSN()
	if !strings.Contains(dsn, "seuoj") || !strings.Contains(dsn, "tcp(localhost:3306)") {
		t.Fatalf("unexpected dsn: %q", dsn)
	}
}

func TestNewRejectsUnreachableMySQL(t *testing.T) {
	_, err := New(config.DatabaseConfig{
		Host: "127.0.0.1", Port: "1", User: "nobody", Password: "bad", Name: "seuoj",
	})
	if err == nil {
		t.Fatal("expected database connection failure")
	}
}
