# Local setup

Monorepo: `backend/` (Go, hexagonal architecture — REST for sprints, gRPC for work items) +
`frontend/` (Next.js).

## Prerequisites

- Go 1.25+ (`go version`)
- Docker + Docker Compose
- Node 20+ and pnpm (`corepack enable` gets you pnpm if missing)
- `goose` (Postgres migration tool) — installed in step 3 below

## 1. Databases

From repo root:

```bash
docker compose up -d db postgres
docker compose ps   # both services should show "healthy"
```

`db` = MongoDB (work items), `postgres` = Postgres (sprints).

Default host ports: Mongo `27017`, Postgres `5434`. If either is already taken by something
else on your machine, remap the host side in `docker-compose.yml` (left number in `"HOST:CONTAINER"`)
and update `DATABASE_URL` in your `.env` to match.

## 2. Backend env

```bash
cd backend
cp .env.example .env
```

Open `.env` and check:

```
DATABASE_URL=postgres://chura:chura@localhost:5434/chura?sslmode=disable
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=chura
FRONTEND_URL=http://localhost:3000
PORT=8080
GRPC_PORT=9000
```

`GRPC_PORT` — check nothing else on your machine already owns 9000:
`lsof -nP -iTCP:9000 -sTCP:LISTEN`. If it's taken, pick a free port (e.g. `19000`) and use that
same value wherever you point Postman/grpcurl at the gRPC server later.

## 3. Run Postgres migrations

```bash
# still in backend/
go install github.com/pressly/goose/v3/cmd/goose@latest
export PATH=$PATH:$(go env GOPATH)/bin

make migrate-up
```

Expect:
```
OK   001_create_sprints.sql
OK   002_add_sprint_timestamps.sql
OK   003_sprint_gorm_model_pk.sql
goose: successfully migrated database to version: 3
```

## 4. Run the backend

```bash
# still in backend/
go run ./cmd/api
```

Expect: `server running on :8080 (grpc :<your GRPC_PORT>)`. Leave this running.

Sanity check in another terminal:

```bash
curl localhost:8080/health
# {"status":"healthy"}
```

## 5. Run the frontend

```bash
cd frontend
pnpm install
pnpm dev
```

Open http://localhost:3000.

## 6. Test the backend

### REST (sprints) — import into Postman

1. Postman → **Import** (top-left) → select
   `backend/docs/postman/Chura-REST.postman_collection.json` → Import.
2. **Import** again → select `backend/docs/postman/Chura.postman_environment.json` → Import.
3. Top-right environment dropdown → pick **Chura - Local**.
4. Backend must be running (step 4 above).
5. Open the collection, run requests top to bottom: Health Check → Create → List → Get By Id →
   Update → Create Invalid (expect 400) → Delete → Get After Delete (expect 404).
   Create must run first — it captures `sprint_id` into a collection variable the rest reuse.
6. Or run the whole thing at once: collection **···** → **Run collection** (Runner executes all
   requests + `pm.test` assertions, shows pass/fail).

Curl equivalent, if you don't want Postman:

```bash
curl -s -X POST localhost:8080/api/v1/sprints -H 'Content-Type: application/json' \
  -d '{"name":"Sprint 1","team":"Platform","start_date":"2026-10-01","end_date":"2026-10-14","status":"planned"}' | tee /tmp/sprint.json

SPRINT_ID=$(python3 -c "import json;print(json.load(open('/tmp/sprint.json'))['id'])")
curl -s localhost:8080/api/v1/sprints/$SPRINT_ID
curl -s -X PUT localhost:8080/api/v1/sprints/$SPRINT_ID -H 'Content-Type: application/json' \
  -d '{"name":"Sprint 1","team":"Platform","status":"active"}'
curl -s -X DELETE localhost:8080/api/v1/sprints/$SPRINT_ID -w '%{http_code}\n'
```

### gRPC (work items) — Postman can't import this one as JSON

Postman has no export/import path for gRPC requests yet (confirmed against their own community
forum). So it's built manually in the app, once, then saved into a collection:

1. Postman → **New** → **gRPC Request**.
2. URL field: `localhost:9000` (or your `GRPC_PORT` value if you changed it — no `grpc://`
   prefix, Postman adds that itself; make sure the field has your real value, not the
   `<value>` placeholder text).
3. **Service definition** tab → **Select .proto file** → browse to
   `backend/internal/adapter/handler/grpc/proto/workitem.proto` → Import.
4. Method dropdown (top, next to URL) → pick a method. Start with `GetWorkItem` (unary, simplest)
   to confirm the connection works before touching the streaming one.
5. **Message** tab → **Use Example Message** to scaffold JSON, edit values.
6. Unary methods (`GetWorkItem`, `ListWorkItems`, `UpdateWorkItem`, `DeleteWorkItem`): click
   **Invoke** — one request, one response.
7. `CreateWorkItems` is client-streaming — different flow:
   - Click **Invoke** (opens the stream, button becomes **Send**)
   - Click **Send** to push the message — server replies per-message, appears in Responses
   - Click **End Streaming** to close cleanly
8. Save each configured request into a new collection (e.g. "Chura - gRPC") so you don't have
   to reconfigure next time — that save works fine, it's only cross-machine JSON import that's
   unsupported.

Scriptable alternative (no manual UI setup, good for repeat runs / CI) — full copy-paste
`grpcurl` script for create/get/list/update/delete: `backend/docs/postman/grpc-workitem.md`.

### Prove it hit the real DB

```bash
# Postgres — run after each REST step above
psql postgres://chura:chura@localhost:5434/chura -c \
  "select id, name, team, status, start_date, end_date, created_at, updated_at, deleted_at from sprints order by id;"

# Mongo — run after each grpcurl/Postman gRPC step
mongosh "mongodb://localhost:27017/chura" --eval "db.work_items.find().pretty()"
```

Full walkthrough with expected output at each step: `backend/docs/testing/manual-testing.md`.

## Common issues

| Symptom | Cause | Fix |
|---|---|---|
| `DATABASE_URL is not set` on startup | no `.env` in `backend/`, or ran from wrong dir | `cp .env.example .env` inside `backend/` |
| `goose: no migrations found` / migrate-up does nothing | old Makefile pointed at wrong dir | already fixed — migrations live in `internal/adapter/db/postgres/migrations` |
| Postman gRPC gets `UNIMPLEMENTED`/weird errors, not a timeout | something else already listening on the port you pointed Postman at | `lsof -nP -iTCP:<port> -sTCP:LISTEN`, set `GRPC_PORT` to a free one, restart backend, repoint Postman |
| Postman gRPC gets `UNAVAILABLE` | nothing listening at all — backend not actually running | check your terminal running `go run ./cmd/api` is still up |
| `bind: address already in use` on `docker compose up` | port 27017/5434 taken by another local service | remap host port in `docker-compose.yml` |
