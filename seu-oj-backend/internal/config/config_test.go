package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigAndFormatting(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Server.Address() != "0.0.0.0:8080" {
		t.Fatalf("unexpected server address %q", cfg.Server.Address())
	}
	if dsn := cfg.Database.DSN(); !strings.Contains(dsn, "root:@tcp(127.0.0.1:3306)/seu_oj") {
		t.Fatalf("unexpected database dsn %q", dsn)
	}
	if cfg.Sandbox.DockerImage != "gcc:13" || !cfg.Sandbox.ReadOnlyRootFS {
		t.Fatalf("unexpected sandbox defaults: %+v", cfg.Sandbox)
	}
}

func TestLoadReadsYAMLAndAppliesEnvironmentOverrides(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	yaml := []byte(`
server:
  host: "127.0.0.1"
  port: "9000"
database:
  host: "db.internal"
  port: "3307"
  user: "oj"
  password: "secret"
  name: "seu_test"
redis:
  addr: "redis.internal:6379"
  password: "redis-secret"
  db: 2
auth:
  jwt_secret: "yaml-secret"
sandbox:
  docker_image: "custom:latest"
  memory_mb: 512
  read_only_rootfs: false
`)
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), yaml, 0644); err != nil {
		t.Fatalf("write config yaml: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	t.Setenv("SERVER_PORT", "9100")
	t.Setenv("REDIS_DB", "7")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("SANDBOX_MEMORY_MB", "1024")
	t.Setenv("SANDBOX_READ_ONLY_ROOTFS", "yes")

	cfg := Load()
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != "9100" {
		t.Fatalf("unexpected server config: %+v", cfg.Server)
	}
	if cfg.Database.DSN() != "oj:secret@tcp(db.internal:3307)/seu_test?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("unexpected dsn: %s", cfg.Database.DSN())
	}
	if cfg.Redis.DB != 7 || cfg.Redis.Addr != "redis.internal:6379" {
		t.Fatalf("unexpected redis config: %+v", cfg.Redis)
	}
	if cfg.Auth.JWTSecret != "env-secret" {
		t.Fatalf("expected env jwt secret, got %q", cfg.Auth.JWTSecret)
	}
	if cfg.Sandbox.MemoryMB != 1024 || !cfg.Sandbox.ReadOnlyRootFS {
		t.Fatalf("unexpected sandbox env overrides: %+v", cfg.Sandbox)
	}
}

func TestEnvironmentParsersFallbackOnInvalidInput(t *testing.T) {
	t.Setenv("INT_VALUE", "not-an-int")
	t.Setenv("BOOL_VALUE", "maybe")

	if got := getEnvAsInt("INT_VALUE", 42); got != 42 {
		t.Fatalf("expected int fallback, got %d", got)
	}
	if got := getEnvAsBool("BOOL_VALUE", true); !got {
		t.Fatal("expected bool fallback true")
	}

	t.Setenv("BOOL_VALUE", "off")
	if got := getEnvAsBool("BOOL_VALUE", true); got {
		t.Fatal("expected false for off")
	}
}
