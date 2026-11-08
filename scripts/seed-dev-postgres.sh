#!/usr/bin/env bash
# Seeds local Postgres + MongoDB + Keycloak membership for Board/Sprints/project UI.
# Requires migrations applied (make migrate) and Keycloak up (make infra).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$REPO_ROOT/backend/.env"
SQL_FILE="$REPO_ROOT/backend/scripts/seed/dev_project.sql"
MONGO_SEED="$REPO_ROOT/backend/scripts/seed/dev_work_items.js"
KC_SEED="$REPO_ROOT/scripts/seed-dev-keycloak.sh"
PG_CONTAINER=chura-agile-execution-postgres
MONGO_CONTAINER=chura-agile-execution-mongodb
PROJECT_ID=11111111-1111-4111-8111-111111111111

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

echo "→ Postgres: project + sprints…"
run_psql -f - < "$SQL_FILE"
echo "✔ Postgres seed applied ($(basename "$SQL_FILE"))"

if [[ -f "$MONGO_SEED" ]]; then
  echo "→ MongoDB: work items…"
  if docker ps --format '{{.Names}}' | grep -qx "$MONGO_CONTAINER"; then
    docker exec -i "$MONGO_CONTAINER" mongosh --quiet < "$MONGO_SEED"
  elif command -v mongosh >/dev/null 2>&1; then
    mongosh --quiet < "$MONGO_SEED"
  else
    echo "⚠️  skipping Mongo seed — MongoDB not reachable"
  fi
else
  echo "⚠️  skipping Mongo seed — missing $MONGO_SEED"
fi

if [[ -f "$KC_SEED" ]]; then
  echo "→ Keycloak: project membership…"
  if bash "$KC_SEED"; then
    :
  else
    echo "⚠️  Keycloak membership seed failed — project row + sprints still applied."
    echo "   Fix Keycloak (make infra / make reset-keycloak) and re-run: make seed-dev"
  fi
else
  echo "⚠️  skipping Keycloak membership — missing $KC_SEED"
fi

echo
echo "✔ seed-dev complete"
echo "  Project id: $PROJECT_ID"
echo "  Set frontend NEXT_PUBLIC_CHURA_PROJECT_ID=$PROJECT_ID (then restart pnpm dev)"
echo "  Users: tonnam / pim (member), jojo (auditor) — password test1234"
