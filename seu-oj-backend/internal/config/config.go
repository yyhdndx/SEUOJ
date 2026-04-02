package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	Sandbox  SandboxConfig  `yaml:"sandbox"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type SandboxConfig struct {
	DockerImage        string `yaml:"docker_image"`
	CompileImage       string `yaml:"compile_image"`
	RunImage           string `yaml:"run_image"`
	User               string `yaml:"user"`
	CPUs               string `yaml:"cpus"`
	MemoryMB           int    `yaml:"memory_mb"`
	PIDsLimit          int    `yaml:"pids_limit"`
	TmpfsMB            int    `yaml:"tmpfs_mb"`
	FileSizeKB         int    `yaml:"file_size_kb"`
	OutputLimitKB      int    `yaml:"output_limit_kb"`
	CompileOutputKB    int    `yaml:"compile_output_kb"`
	ReadOnlyRootFS     bool   `yaml:"read_only_rootfs"`
	CompileTimeoutSec  int    `yaml:"compile_timeout_sec"`
	RunTimeoutBufferMS int    `yaml:"run_timeout_buffer_ms"`
}

func Load() Config {
	cfg := defaultConfig()

	loadFromYAML(&cfg, "config/config.yaml")
	overrideFromEnv(&cfg)

	return cfg
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     "3306",
			User:     "root",
			Password: "",
			Name:     "seu_oj",
		},
		Redis: RedisConfig{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		},
		Auth: AuthConfig{
			JWTSecret: "seu-oj-dev-secret",
		},
		Sandbox: SandboxConfig{
			DockerImage:        "gcc:13",
			CompileImage:       "",
			RunImage:           "",
			User:               "65534:65534",
			CPUs:               "1.0",
			MemoryMB:           256,
			PIDsLimit:          64,
			TmpfsMB:            64,
			FileSizeKB:         1024,
			OutputLimitKB:      256,
			CompileOutputKB:    256,
			ReadOnlyRootFS:     true,
			CompileTimeoutSec:  15,
			RunTimeoutBufferMS: 500,
		},
	}
}

func loadFromYAML(cfg *Config, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	_ = yaml.Unmarshal(content, cfg)
}

func overrideFromEnv(cfg *Config) {
	cfg.Server.Host = getEnv("SERVER_HOST", cfg.Server.Host)
	cfg.Server.Port = getEnv("SERVER_PORT", cfg.Server.Port)
	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnv("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnv("DB_NAME", cfg.Database.Name)
	cfg.Redis.Addr = getEnv("REDIS_ADDR", cfg.Redis.Addr)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", cfg.Redis.DB)
	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Sandbox.DockerImage = getEnv("SANDBOX_DOCKER_IMAGE", cfg.Sandbox.DockerImage)
	cfg.Sandbox.CompileImage = getEnv("SANDBOX_COMPILE_IMAGE", cfg.Sandbox.CompileImage)
	cfg.Sandbox.RunImage = getEnv("SANDBOX_RUN_IMAGE", cfg.Sandbox.RunImage)
	cfg.Sandbox.User = getEnv("SANDBOX_USER", cfg.Sandbox.User)
	cfg.Sandbox.CPUs = getEnv("SANDBOX_CPUS", cfg.Sandbox.CPUs)
	cfg.Sandbox.MemoryMB = getEnvAsInt("SANDBOX_MEMORY_MB", cfg.Sandbox.MemoryMB)
	cfg.Sandbox.PIDsLimit = getEnvAsInt("SANDBOX_PIDS_LIMIT", cfg.Sandbox.PIDsLimit)
	cfg.Sandbox.TmpfsMB = getEnvAsInt("SANDBOX_TMPFS_MB", cfg.Sandbox.TmpfsMB)
	cfg.Sandbox.FileSizeKB = getEnvAsInt("SANDBOX_FILE_SIZE_KB", cfg.Sandbox.FileSizeKB)
	cfg.Sandbox.OutputLimitKB = getEnvAsInt("SANDBOX_OUTPUT_LIMIT_KB", cfg.Sandbox.OutputLimitKB)
	cfg.Sandbox.CompileOutputKB = getEnvAsInt("SANDBOX_COMPILE_OUTPUT_KB", cfg.Sandbox.CompileOutputKB)
	cfg.Sandbox.ReadOnlyRootFS = getEnvAsBool("SANDBOX_READ_ONLY_ROOTFS", cfg.Sandbox.ReadOnlyRootFS)
	cfg.Sandbox.CompileTimeoutSec = getEnvAsInt("SANDBOX_COMPILE_TIMEOUT_SEC", cfg.Sandbox.CompileTimeoutSec)
	cfg.Sandbox.RunTimeoutBufferMS = getEnvAsInt("SANDBOX_RUN_TIMEOUT_BUFFER_MS", cfg.Sandbox.RunTimeoutBufferMS)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return fallback
	}
	return result
}

func getEnvAsBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
