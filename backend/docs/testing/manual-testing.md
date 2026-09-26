# Manual testing guide

Covers: bring up both DBs, run the API, exercise REST (sprint) + gRPC (workitem), then prove the
rows actually changed in Postgres/Mongo.

## 1. Bring up the databases

From repo root:

```bash
docker compose up -d db postgres
```

`db` = Mongo (workitems), `postgres` = Postgres (sprints). Wait for both healthy:
`docker compose ps`.

## 2. Configure and migrate

```bash
cd backend
cp .env.example .env
# .env already points at localhost:5432 / localhost:27017, matching the compose ports above

go install github.com/pressly/goose/v3/cmd/goose@latest   # once, if not installed
make migrate-up
```

This runs `internal/adapter/db/postgres/migrations/*.sql`, creating the `sprints` table.

## 3. Run the API

```bash
go run ./cmd/api
```

Expect: `server running on :8080`. REST listens on `:8080`, gRPC on `:9000`.

## 4. REST test set (Sprint API)

Import into Postman:
- `docs/postman/Chura-REST.postman_collection.json`
- `docs/postman/Chura.postman_environment.json`

Run the collection top to bottom (Runner or manually) — each request has `pm.test` assertions
and the Create step feeds `sprint_id` into the rest automatically.

Or curl, same flow:

```bash
curl -s -X POST localhost:8080/api/v1/sprints \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sprint 1","team":"Platform","start_date":"2026-10-01","end_date":"2026-10-14","status":"planned"}' | tee /tmp/sprint.json

SPRINT_ID=$(python3 -c "import json;print(json.load(open('/tmp/sprint.json'))['id'])")
curl -s localhost:8080/api/v1/sprints/$SPRINT_ID
curl -s -X PUT localhost:8080/api/v1/sprints/$SPRINT_ID \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sprint 1","team":"Platform","status":"active"}'
curl -s -X DELETE localhost:8080/api/v1/sprints/$SPRINT_ID -w '%{http_code}\n'
```

## 5. gRPC test set (WorkItem service)

See [`docs/postman/grpc-workitem.md`](../postman/grpc-workitem.md) — Postman manual gRPC setup
+ full `grpcurl` script (create/get/list/update/delete).

## 6. Prove it in the real DB

**Postgres (sprints) — while the REST requests above are running, in another terminal:**

```bash
psql postgres://chura:chura@localhost:5434/chura -c \
  "select id, name, team, status, start_date, end_date, created_at, updated_at, deleted_at from sprints order by id;"
```

Run it:
- right after Create → new row, `status='planned'`, `deleted_at` NULL
- right after Update → same row, `status='active'`, `updated_at` bumped
- right after Delete → row still physically present but `deleted_at` set (GORM soft-delete);
  the API returns 404 for it because GORM's default scope filters `deleted_at IS NULL`

**Mongo (work_items) — while running the grpcurl script:**

```bash
mongosh "mongodb://localhost:27017/chura" --eval \
  "db.work_items.find().pretty()"
```

Run it after each grpcurl step (create/update/delete) and watch the document appear, change
`status`, then disappear.

## Notes

- Sprint IDs are Postgres `BIGSERIAL` (`1`, `2`, ...) — not UUIDs, despite the `json:"id"` string
  type; that's why the migrations dropped the original UUID PK (see
  `003_sprint_gorm_model_pk.sql`).
- WorkItem IDs are UUIDv4, generated in `internal/service/workitem.go`.
- `sslmode=disable` in `DATABASE_URL` is fine for local Docker Postgres; don't reuse that flag
  against a real deployment.
