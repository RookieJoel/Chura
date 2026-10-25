#!/usr/bin/env bash
# Run all four domain services locally (no Kong). Paths are per-port; use `make up` for a unified gateway on :8083.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/backend"

trap 'kill 0' EXIT INT TERM

export CHURA_ENV_FILE="${CHURA_ENV_FILE:-.env}"

run() {
  local name="$1" port="$2" consul_name="$3"
  PORT="$port" CONSUL_SERVICE_NAME="$consul_name" CONSUL_SERVICE_PORT="$port" \
    go run "./cmd/$name" &
}

run analytics 8056 chura-analytics
run sprint 8057 chura-sprint
run workitem 8058 chura-workitem
run project 8059 chura-project

wait
