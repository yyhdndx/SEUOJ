package router

import (
	"testing"

	"seu-oj-backend/internal/config"
)

func TestNewRegistersRoutes(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Auth: config.AuthConfig{JWTSecret: "test-secret"},
		Sandbox: config.SandboxConfig{
			DockerImage: "gcc:13",
			User:        "65534:65534",
			CPUs:        "1.0",
			MemoryMB:    256,
			PIDsLimit:   64,
			TmpfsMB:     64,
		},
	}

	engine := New(nil, nil, cfg)
	if engine == nil {
		t.Fatal("expected router engine")
	}
}
