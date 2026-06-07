#!/usr/bin/env bash
# Read seu-oj-backend/config/config.yaml without external yaml tools.

set -euo pipefail

yaml_get() {
  local file="$1" section="$2" key="$3"
  awk -v section="$section" -v key="$key" '
    BEGIN { in_section = 0 }
    $0 ~ "^" section ":" { in_section = 1; next }
    in_section && /^[^[:space:]#]/ { in_section = 0 }
    in_section && $1 == key":" {
      val = $0
      sub(/^[^:]*:[[:space:]]*/, "", val)
      gsub(/^["'\''"]|["'\''"]$/, "", val)
      print val
      exit
    }
  ' "$file"
}

resolve_config_path() {
  local backend_dir="$1"
  if [[ -f "$backend_dir/config/config.yaml" ]]; then
    echo "$backend_dir/config/config.yaml"
  elif [[ -f "$backend_dir/config/config.example.yaml" ]]; then
    echo "$backend_dir/config/config.example.yaml"
  else
    echo "missing config under $backend_dir" >&2
    return 1
  fi
}

load_database_config() {
  local config_file="$1"

  DB_HOST="${DB_HOST:-$(yaml_get "$config_file" database host)}"
  DB_PORT="${DB_PORT:-$(yaml_get "$config_file" database port)}"
  DB_USER="${DB_USER:-$(yaml_get "$config_file" database user)}"
  DB_PASSWORD="${DB_PASSWORD:-$(yaml_get "$config_file" database password)}"
  DB_NAME="${DB_NAME:-$(yaml_get "$config_file" database name)}"

  DB_HOST="${DB_HOST:-127.0.0.1}"
  DB_PORT="${DB_PORT:-3306}"
  DB_USER="${DB_USER:-root}"
  DB_PASSWORD="${DB_PASSWORD:-}"
  DB_NAME="${DB_NAME:-seu_oj}"
}

load_redis_config() {
  local config_file="$1"
  local addr
  addr="$(yaml_get "$config_file" redis addr)"
  REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
  REDIS_PORT="${REDIS_PORT:-6379}"
  if [[ "$addr" =~ ^(.+):([0-9]+)$ ]]; then
    REDIS_HOST="${BASH_REMATCH[1]}"
    REDIS_PORT="${BASH_REMATCH[2]}"
  fi
}

mysql_cmd() {
  if [[ -n "${MYSQL_BIN:-}" ]]; then
    echo "$MYSQL_BIN"
  elif command -v mysql >/dev/null 2>&1; then
    command -v mysql
  else
    echo "mysql"
  fi
}

run_mysql() {
  local db_name="${1:-}"
  shift || true
  local bin
  bin="$(mysql_cmd)"
  if [[ -n "$DB_PASSWORD" ]]; then
    MYSQL_PWD="$DB_PASSWORD" "$bin" -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" ${db_name:+"$db_name"} "$@"
  else
    "$bin" -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" ${db_name:+"$db_name"} "$@"
  fi
}
