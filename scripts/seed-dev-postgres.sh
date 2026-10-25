#!/usr/bin/env bash
# Seeds local Postgres dev data. Requires migrations applied (make migrate).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$REPO_ROOT/backend/.env"
SQL_FILE="$REPO_ROOT/backend/scripts/seed/dev_project.sql"
PG_CONTAINER=chura-agile-execution-postgres

if [[ ! -f "$ENV_FILE" ]]; then
  echo "❌ missing $ENV_FILE — run: make env"
  exit 1
fi

if [[ ! -f "$SQL_FILE" ]]; then
  echo "❌ missing seed file: $SQL_FILE"
  exit 1
fi

DATABASE_URL="$(grep -E '^DATABASE_URL=' "$ENV_FILE" | head -1 | cut -d= -f2- | tr -d '"' | tr -d "'")"
if [[ -z "$DATABASE_URL" ]]; then
  echo "❌ DATABASE_URL is not set in $ENV_FILE"
  exit 1
fi

run_psql() {
  if docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER"; then
    docker exec -i "$PG_CONTAINER" psql -U chura -d chura -v ON_ERROR_STOP=1 "$@"
    return
  fi
  if command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 "$@"
    return
  fi
  echo "❌ Postgres is not reachable: start infra (make infra) or install psql."
  exit 1
}

if ! run_psql -c "SELECT 1 FROM projects LIMIT 1" >/dev/null 2>&1; then
  echo "❌ table projects is missing — run: make migrate"
  exit 1
fi

run_psql -f - < "$SQL_FILE"
echo "✔ dev Postgres seed applied ($(basename "$SQL_FILE"))"
