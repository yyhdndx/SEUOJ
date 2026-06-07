package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"seu-oj-backend/internal/config"
)

var (
	schemaFiles = []string{
		"user.sql",
		"problems.sql",
		"problem_testcases.sql",
		"contests.sql",
		"contest_problems.sql",
		"contest_registrations.sql",
		"contest_announcements.sql",
		"submissions.sql",
		"submission_results.sql",
		"forum_topics.sql",
		"forum_replies.sql",
		"audit_logs.sql",
	}
	seedFiles = []string{
		"seed_more_problems.sql",
		"seed_classes_assignments.sql",
		"seed_contests.sql",
		"seed_forum.sql",
		"seed_playlists.sql",
		"seed_problem_solutions.sql",
	}
)

func main() {
	check := flag.Bool("check", false, "exit 0 when database is initialized")
	ping := flag.Bool("ping", false, "exit 0 when MySQL is reachable")
	force := flag.Bool("force", false, "re-apply schema and seed even if already initialized")
	flag.Parse()

	cfg := config.Load()
	dbDir := resolveDatabaseDir()

	if *ping {
		if err := pingMySQL(cfg.Database); err != nil {
			fmt.Fprintf(os.Stderr, "mysql ping failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("mysql ok")
		return
	}

	if *check {
		ok, err := isInitialized(cfg.Database)
		if err != nil {
			fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
			os.Exit(1)
		}
		if !ok {
			os.Exit(1)
		}
		fmt.Println("database initialized")
		return
	}

	if !*force {
		ok, err := isInitialized(cfg.Database)
		if err != nil {
			log.Fatalf("check database: %v", err)
		}
		if ok {
			fmt.Printf("database %q already initialized, skip (use --force to re-apply)\n", cfg.Database.Name)
			return
		}
	}

	fmt.Printf("using config database: host=%s port=%s user=%s name=%s\n",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name)

	if err := runInit(cfg.Database, dbDir); err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	fmt.Printf("database init completed for %q\n", cfg.Database.Name)
}

func resolveDatabaseDir() string {
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, "database")
		if dirExists(candidate) {
			return candidate
		}
		parent := filepath.Join(wd, "..", "database")
		if dirExists(parent) {
			return parent
		}
	}
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "database")
		if dirExists(candidate) {
			return candidate
		}
	}
	log.Fatal("cannot locate database SQL directory")
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func dsn(cfg config.DatabaseConfig, withDB bool) string {
	dbName := ""
	if withDB {
		dbName = cfg.Name
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		dbName,
	)
}

func openDB(cfg config.DatabaseConfig, withDB bool) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn(cfg, withDB))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func pingMySQL(cfg config.DatabaseConfig) error {
	db, err := openDB(cfg, false)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("SELECT 1")
	return err
}

func isInitialized(cfg config.DatabaseConfig) (bool, error) {
	admin, err := openDB(cfg, false)
	if err != nil {
		return false, err
	}
	defer admin.Close()

	var tableCount int
	err = admin.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'users'",
		cfg.Name,
	).Scan(&tableCount)
	if err != nil {
		return false, err
	}
	if tableCount != 1 {
		return false, nil
	}

	db, err := openDB(cfg, true)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		return false, err
	}
	return userCount > 0, nil
}

func runInit(cfg config.DatabaseConfig, dbDir string) error {
	admin, err := openDB(cfg, false)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}
	defer admin.Close()

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Name,
	)
	if _, err := admin.Exec(createSQL); err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	db, err := openDB(cfg, true)
	if err != nil {
		return fmt.Errorf("connect database %q: %w", cfg.Name, err)
	}
	defer db.Close()

	for _, file := range schemaFiles {
		if err := applySQLFile(db, filepath.Join(dbDir, file)); err != nil {
			return err
		}
	}
	for _, file := range seedFiles {
		if err := applySQLFile(db, filepath.Join(dbDir, file)); err != nil {
			return err
		}
	}
	if err := applyIndexSQL(db, filepath.Join(dbDir, "index.sql")); err != nil {
		return err
	}
	return nil
}

func applySQLFile(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	fmt.Printf("applying %s ...\n", filepath.Base(path))
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("apply %s: %w", path, err)
	}
	return nil
}

var callLinePattern = regexp.MustCompile(`(?m)^CALL\s+.+;\s*$`)

func applyIndexSQL(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	procedureSQL := extractCreateProcedure(string(content))
	if procedureSQL == "" {
		return fmt.Errorf("parse procedure from %s", path)
	}

	fmt.Printf("applying %s ...\n", filepath.Base(path))
	if _, err := db.Exec("DROP PROCEDURE IF EXISTS add_index_if_missing"); err != nil {
		return fmt.Errorf("drop index helper procedure: %w", err)
	}
	if _, err := db.Exec(procedureSQL); err != nil {
		return fmt.Errorf("create index helper procedure: %w", err)
	}

	for _, line := range callLinePattern.FindAllString(string(content), -1) {
		line = strings.TrimSpace(line)
		if _, err := db.Exec(line); err != nil {
			return fmt.Errorf("apply index statement %q: %w", line, err)
		}
	}

	if _, err := db.Exec("DROP PROCEDURE IF EXISTS add_index_if_missing"); err != nil {
		return fmt.Errorf("cleanup index helper procedure: %w", err)
	}
	return nil
}

func extractCreateProcedure(content string) string {
	start := strings.Index(content, "CREATE PROCEDURE add_index_if_missing")
	if start < 0 {
		return ""
	}
	end := strings.Index(content[start:], "END//")
	if end < 0 {
		end = strings.Index(content[start:], "END;")
		if end < 0 {
			return ""
		}
		return strings.TrimSpace(content[start : start+end+3])
	}
	return strings.TrimSpace(content[start : start+end+3])
}
