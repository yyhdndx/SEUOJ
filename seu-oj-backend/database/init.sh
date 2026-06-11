#!/usr/bin/env bash
# Initialize SEU OJ MySQL schema and demo data via Go (no mysql CLI required).
# Reads ../config/config.yaml automatically.
#
# Usage:
#   cd seu-oj-backend/database
#   bash init.sh
#   bash init.sh --force

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"
exec go run ./cmd/db-init "$@"
