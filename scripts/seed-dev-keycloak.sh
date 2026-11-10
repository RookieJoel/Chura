#!/usr/bin/env bash
# Wire Keycloak membership for the seeded Postgres project.
# Uses the Keycloak admin user (local only). Idempotent.
# Requires: curl, node.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$REPO_ROOT/backend/.env"
PROJECT_ID="${PROJECT_ID:-11111111-1111-4111-8111-111111111111}"
GROUP_NAME="project-${PROJECT_ID}"
KC_URL="${KC_URL:-http://localhost:8080}"
KC_REALM="${KC_REALM:-chura}"
KC_ADMIN_USER="${KC_ADMIN_USER:-admin}"
KC_ADMIN_PASSWORD="${KC_ADMIN_PASSWORD:-admin}"
PG_CONTAINER=chura-agile-execution-postgres
TMPDIR_SEED="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_SEED"' EXIT

# username|email|project_role
MEMBERS=(
  "tonnam|tonnam@example.com|product_owner"
  "pim|pim@example.com|developer"
)

if [[ ! -f "$ENV_FILE" ]]; then
  echo "❌ missing $ENV_FILE — run: make env"
  exit 1
fi

DATABASE_URL="$(grep -E '^DATABASE_URL=' "$ENV_FILE" | head -1 | cut -d= -f2- | tr -d '"' | tr -d "'")"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "❌ missing required command: $1"
    exit 1
  }
}
need_cmd curl
need_cmd node

node_json() {
  local file="$1"
  local expr="$2"
  node --input-type=module -e "
import { readFileSync } from 'node:fs';
const d = JSON.parse(readFileSync(process.argv[1], 'utf8'));
const v = ($expr);
process.stdout.write(v == null ? '' : String(v));
" "$file"
}

run_psql() {
  if docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER"; then
    docker exec -i "$PG_CONTAINER" psql -U chura -d chura -v ON_ERROR_STOP=1 -t -A "$@"
    return
  fi
  if command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -t -A "$@"
    return
  fi
  echo "❌ Postgres is not reachable: start infra (make infra) or install psql."
  exit 1
}

echo "→ Keycloak login as ${KC_ADMIN_USER}…"
curl -sf -X POST "$KC_URL/realms/master/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=$KC_ADMIN_USER" \
  -d "password=$KC_ADMIN_PASSWORD" \
  >"$TMPDIR_SEED/token.json"
TOKEN="$(node_json "$TMPDIR_SEED/token.json" 'd.access_token')"
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "❌ could not get Keycloak admin token — is Keycloak up? (make infra)"
  exit 1
fi

auth=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")

echo "→ ensure group ${GROUP_NAME}…"
curl -sf "${auth[@]}" \
  --get "$KC_URL/admin/realms/$KC_REALM/groups" \
  --data-urlencode "search=$GROUP_NAME" \
  --data-urlencode "exact=true" \
  >"$TMPDIR_SEED/groups.json"
GROUP_ID="$(node --input-type=module -e "
import { readFileSync } from 'node:fs';
const groups = JSON.parse(readFileSync(process.argv[1], 'utf8'));
const name = process.argv[2];
const hit = groups.find((g) => g && g.name === name);
process.stdout.write(hit && hit.id ? hit.id : '');
" "$TMPDIR_SEED/groups.json" "$GROUP_NAME")"

if [[ -z "$GROUP_ID" ]]; then
  curl -sf "${auth[@]}" -X POST "$KC_URL/admin/realms/$KC_REALM/groups" \
    -d "{\"name\":\"$GROUP_NAME\"}" >/dev/null
  curl -sf "${auth[@]}" \
    --get "$KC_URL/admin/realms/$KC_REALM/groups" \
    --data-urlencode "search=$GROUP_NAME" \
    --data-urlencode "exact=true" \
    >"$TMPDIR_SEED/groups.json"
  GROUP_ID="$(node --input-type=module -e "
import { readFileSync } from 'node:fs';
const groups = JSON.parse(readFileSync(process.argv[1], 'utf8'));
const name = process.argv[2];
const hit = groups.find((g) => g && g.name === name);
process.stdout.write(hit && hit.id ? hit.id : '');
" "$TMPDIR_SEED/groups.json" "$GROUP_NAME")"
fi

if [[ -z "$GROUP_ID" ]]; then
  echo "❌ failed to create or find Keycloak group $GROUP_NAME"
  exit 1
fi

echo "→ point projects.keycloak_group_id at $GROUP_ID…"
run_psql -c "UPDATE projects SET keycloak_group_id = '$GROUP_ID', updated_at = NOW() WHERE id = '$PROJECT_ID';"

assign_member() {
  local username="$1" email="$2" role="$3"
  local user_id

  curl -sf "${auth[@]}" \
    --get "$KC_URL/admin/realms/$KC_REALM/users" \
    --data-urlencode "username=$username" \
    --data-urlencode "exact=true" \
    >"$TMPDIR_SEED/users.json"
  user_id="$(node_json "$TMPDIR_SEED/users.json" 'd[0] && d[0].id')"
  if [[ -z "$user_id" ]]; then
    echo "⚠️  skip $username — user not found in realm $KC_REALM"
    return 0
  fi

  curl -sf "${auth[@]}" -X PUT \
    "$KC_URL/admin/realms/$KC_REALM/users/$user_id/groups/$GROUP_ID" >/dev/null

  curl -sf "${auth[@]}" "$KC_URL/admin/realms/$KC_REALM/users/$user_id" \
    >"$TMPDIR_SEED/user.json"
  node --input-type=module -e "
import { readFileSync, writeFileSync } from 'node:fs';
const user = JSON.parse(readFileSync(process.argv[1], 'utf8'));
const projectId = process.argv[2];
const role = process.argv[3];
user.attributes = user.attributes || {};
user.attributes['chura_project_role_' + projectId] = [role];
writeFileSync(process.argv[4], JSON.stringify(user));
" "$TMPDIR_SEED/user.json" "$PROJECT_ID" "$role" "$TMPDIR_SEED/user-updated.json"

  curl -sf "${auth[@]}" -X PUT "$KC_URL/admin/realms/$KC_REALM/users/$user_id" \
    --data-binary @"$TMPDIR_SEED/user-updated.json" >/dev/null

  echo "  ✔ $username ($email) → $role"
}

echo "→ add members…"
for entry in "${MEMBERS[@]}"; do
  IFS='|' read -r username email role <<<"$entry"
  assign_member "$username" "$email" "$role"
done

echo "✔ Keycloak membership seeded for project $PROJECT_ID"
echo "  login as tonnam / test1234 to see Chura (dev) in the selector"
