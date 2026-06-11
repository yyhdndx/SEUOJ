#!/usr/bin/env bash
# SEU OJ one-shot dev launcher (Git Bash / WSL / Linux / macOS)
#
# Usage (from repo root):
#   bash scripts/dev.sh
#   bash scripts/dev.sh --setup-only
#   bash scripts/dev.sh --force-db-init
#
# MySQL settings default to seu-oj-backend/config/config.yaml

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$REPO_ROOT/seu-oj-backend"
CODEMIRROR_DIR="$REPO_ROOT/seu-oj-frontend/CodeMirror"
CODEMIRROR_MARKER="$CODEMIRROR_DIR/node_modules/codemirror/dist/index.js"
LOG_DIR="$REPO_ROOT/logs"
WEB_PID_FILE="$LOG_DIR/web.pid"
WORKER_PID_FILE="$LOG_DIR/worker.pid"

SETUP_ONLY=0
FORCE_DB_INIT=0
SKIP_DOCKER_PULL=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --setup-only) SETUP_ONLY=1; shift ;;
    --force-db-init) FORCE_DB_INIT=1; shift ;;
    --skip-docker-pull) SKIP_DOCKER_PULL=1; shift ;;
    -h|--help)
      echo "Usage: bash scripts/dev.sh [--setup-only] [--force-db-init] [--skip-docker-pull]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

# shellcheck source=lib/config.sh
source "$SCRIPT_DIR/lib/config.sh"

step() { echo; echo "==> $*"; }
ok() { echo "    [OK] $*"; }
skip() { echo "    [SKIP] $*"; }
warn() { echo "    [WARN] $*"; }
fail() { echo "    [FAIL] $*"; }

ensure_config_file() {
  step "Check backend config"
  local example="$BACKEND_DIR/config/config.example.yaml"
  local config="$BACKEND_DIR/config/config.yaml"
  if [[ ! -f "$example" ]]; then
    echo "Missing $example" >&2
    exit 1
  fi
  if [[ ! -f "$config" ]]; then
    cp "$example" "$config"
    ok "Created config/config.yaml from template"
    warn "Edit config/config.yaml before demo"
  else
    skip "config/config.yaml already exists"
  fi
}

print_database_settings() {
  step "Database settings (from config.yaml unless overridden)"
  echo "    host: $DB_HOST"
  echo "    port: $DB_PORT"
  echo "    user: $DB_USER"
  echo "    name: $DB_NAME"
  echo "    source: $CONFIG_FILE"
  echo "    db-init: go run ./cmd/db-init (no mysql CLI required)"
}

ensure_database() {
  step "Check MySQL database"
  if ! command -v go >/dev/null 2>&1; then
    echo "go command not found" >&2
    exit 1
  fi

  local needs_init=0
  if [[ "$FORCE_DB_INIT" -eq 1 ]]; then
    needs_init=1
  elif ! (cd "$BACKEND_DIR" && go run ./cmd/db-init --check >/dev/null 2>&1); then
    needs_init=1
  fi

  if [[ "$needs_init" -eq 0 ]]; then
    skip "Database '$DB_NAME' already initialized"
    return
  fi

  if [[ "$FORCE_DB_INIT" -eq 1 ]]; then
    echo "    Force re-initializing database..."
    (cd "$BACKEND_DIR" && go run ./cmd/db-init --force)
  else
    echo "    Database not initialized, importing schema and seed via Go..."
    (cd "$BACKEND_DIR" && go run ./cmd/db-init)
  fi
  ok "Database initialization completed"
}

ensure_codemirror() {
  step "Check CodeMirror frontend deps"
  if [[ ! -f "$CODEMIRROR_DIR/package.json" ]]; then
    warn "CodeMirror/package.json not found, skip npm install"
    return
  fi
  if [[ -f "$CODEMIRROR_MARKER" ]]; then
    skip "CodeMirror node_modules ready"
    return
  fi
  if ! command -v npm >/dev/null 2>&1; then
    echo "npm not found" >&2
    exit 1
  fi
  echo "    Running npm install (first run may take a while)..."
  (cd "$CODEMIRROR_DIR" && npm install --no-fund --no-audit)
  ok "CodeMirror deps installed"
}

check_external_services() {
  step "Check external services"

  if (cd "$BACKEND_DIR" && go run ./cmd/db-init --ping >/dev/null 2>&1); then
    ok "MySQL reachable ($DB_HOST:$DB_PORT/$DB_NAME)"
  else
    fail "MySQL unreachable"
    warn "Verify database settings in config/config.yaml"
  fi

  if command -v redis-cli >/dev/null 2>&1; then
    if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping 2>/dev/null | grep -q PONG; then
      ok "Redis reachable ($REDIS_HOST:$REDIS_PORT)"
    else
      warn "Redis did not respond PONG"
    fi
  else
    warn "redis-cli not found; skipping Redis ping"
  fi

  if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
      ok "Docker available"
      if [[ "$SKIP_DOCKER_PULL" -eq 0 ]]; then
        if docker image inspect gcc:13 >/dev/null 2>&1; then
          skip "gcc:13 image already exists"
        else
          echo "    Pulling judge image gcc:13 ..."
          if docker pull gcc:13 >/dev/null; then
            ok "gcc:13 image ready"
          else
            warn "Failed to pull gcc:13"
          fi
        fi
      fi
    else
      warn "Docker is not running"
    fi
  else
    warn "docker command not found"
  fi

  if ! command -v go >/dev/null 2>&1; then
    echo "go command not found" >&2
    exit 1
  fi
  ok "Go installed: $(go version)"
}

start_dev_processes() {
  step "Start dev services (background + log files)"
  mkdir -p "$LOG_DIR"

  if [[ -f "$WEB_PID_FILE" ]] && kill -0 "$(cat "$WEB_PID_FILE")" 2>/dev/null; then
    warn "Web process already running (PID $(cat "$WEB_PID_FILE")). Run bash scripts/stop-dev.sh first."
  else
    (
      cd "$BACKEND_DIR"
      nohup go run . > "$LOG_DIR/web.log" 2>&1 &
      echo $! > "$WEB_PID_FILE"
    )
    ok "Web server started (PID $(cat "$WEB_PID_FILE"), log: logs/web.log)"
  fi

  if [[ -f "$WORKER_PID_FILE" ]] && kill -0 "$(cat "$WORKER_PID_FILE")" 2>/dev/null; then
    warn "Worker already running (PID $(cat "$WORKER_PID_FILE")). Run bash scripts/stop-dev.sh first."
  else
    (
      cd "$BACKEND_DIR"
      nohup go run ./cmd/judge-worker > "$LOG_DIR/worker.log" 2>&1 &
      echo $! > "$WORKER_PID_FILE"
    )
    ok "Judge worker started (PID $(cat "$WORKER_PID_FILE"), log: logs/worker.log)"
  fi

  echo
  echo "Open in browser: http://127.0.0.1:8080/"
  echo "Tail logs: tail -f logs/web.log logs/worker.log"
  echo "Stop: bash scripts/stop-dev.sh"
}

echo
echo "========================================"
echo "  SEU OJ dev.sh"
echo "========================================"

ensure_config_file
CONFIG_FILE="$(resolve_config_path "$BACKEND_DIR")"
load_database_config "$CONFIG_FILE"
load_redis_config "$CONFIG_FILE"
print_database_settings
check_external_services
ensure_database
ensure_codemirror

if [[ "$SETUP_ONLY" -eq 1 ]]; then
  step "SetupOnly mode"
  ok "Setup finished. Run: bash scripts/dev.sh"
  exit 0
fi

start_dev_processes
