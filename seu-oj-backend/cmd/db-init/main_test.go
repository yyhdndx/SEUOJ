package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"seu-oj-backend/internal/config"
)

func TestExtractCreateProcedureFromIndexSQL(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "database", "index.sql"))
	if err != nil {
		t.Fatalf("read index.sql: %v", err)
	}

	procedure := extractCreateProcedure(string(content))
	if procedure == "" {
		t.Fatal("expected procedure sql")
	}
	if !strings.Contains(procedure, "CREATE PROCEDURE add_index_if_missing") {
		t.Fatal("missing procedure header")
	}
	if !strings.HasSuffix(strings.TrimSpace(procedure), "END") {
		t.Fatalf("expected procedure to end with END, got %q", procedure[len(procedure)-20:])
	}
}

func TestCallLinePatternMatchesIndexSQL(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "database", "index.sql"))
	if err != nil {
		t.Fatalf("read index.sql: %v", err)
	}

	calls := callLinePattern.FindAllString(string(content), -1)
	if len(calls) < 10 {
		t.Fatalf("expected many CALL lines, got %d", len(calls))
	}
	for _, call := range calls {
		if !strings.HasPrefix(strings.TrimSpace(call), "CALL add_index_if_missing(") {
			t.Fatalf("unexpected call line: %q", call)
		}
	}
}

func TestDirExists(t *testing.T) {
	if !dirExists(".") {
		t.Fatal("expected current directory to exist")
	}
	if dirExists("definitely-missing-dir-" + t.Name()) {
		t.Fatal("did not expect missing directory")
	}
}

func TestResolveDatabaseDirFromBackend(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := filepath.Join(wd, "..", "..", "database")
	if !dirExists(dir) {
		t.Skip("database directory not found from test cwd")
	}
}

func TestResolveDatabaseDirFromRepoRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..")
	if !dirExists(filepath.Join(root, "database")) {
		t.Skip("database directory not found")
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir repo root: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	dir := resolveDatabaseDir()
	if !dirExists(dir) {
		t.Fatalf("unexpected database dir %q", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "user.sql")); err != nil {
		t.Fatalf("expected user.sql under %q: %v", dir, err)
	}
}

func TestPingMySQLOptional(t *testing.T) {
	cfg := config.Load()
	if err := pingMySQL(cfg.Database); err != nil {
		t.Skipf("mysql not reachable: %v", err)
	}
}

func TestIsInitializedOptional(t *testing.T) {
	cfg := config.Load()
	ok, err := isInitialized(cfg.Database)
	if err != nil {
		t.Skipf("mysql not reachable: %v", err)
	}
	t.Logf("database initialized=%v", ok)
}

func TestDSNBuildsMySQLConnectionString(t *testing.T) {
	cfg := config.DatabaseConfig{
		User: "root", Password: "secret", Host: "127.0.0.1", Port: "3306", Name: "seuoj",
	}
	withDB := dsn(cfg, true)
	if !strings.Contains(withDB, "/seuoj?") {
		t.Fatalf("expected db name in dsn, got %q", withDB)
	}
	withoutDB := dsn(cfg, false)
	if !strings.Contains(withoutDB, "/?") || strings.Contains(withoutDB, "/seuoj") {
		t.Fatalf("expected empty db segment, got %q", withoutDB)
	}
}

func TestExtractCreateProcedureEmptyInput(t *testing.T) {
	if extractCreateProcedure("") != "" {
		t.Fatal("expected empty procedure for empty input")
	}
}

func TestApplySQLFileRejectsMissingFile(t *testing.T) {
	cfg := config.DatabaseConfig{User: "u", Password: "p", Host: "127.0.0.1", Port: "3306", Name: "db"}
	db, err := openDB(cfg, false)
	if err != nil {
		t.Skipf("mysql not available for applySQLFile test: %v", err)
	}
	defer db.Close()
	if err := applySQLFile(db, filepath.Join(t.TempDir(), "missing.sql")); err == nil {
		t.Fatal("expected missing sql file error")
	}
}

func TestApplySQLFileExecutesSimpleStatement(t *testing.T) {
	cfg := config.Load()
	db, err := openDB(cfg.Database, true)
	if err != nil {
		t.Skipf("mysql not available: %v", err)
	}
	defer db.Close()

	path := filepath.Join(t.TempDir(), "noop.sql")
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write sql: %v", err)
	}
	if err := applySQLFile(db, path); err != nil {
		t.Fatalf("apply sql file: %v", err)
	}
}
