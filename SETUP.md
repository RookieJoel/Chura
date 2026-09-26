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

- REST (sprints): import `backend/docs/postman/Chura-REST.postman_collection.json` +
  `Chura.postman_environment.json` into Postman, or use curl — see
  `backend/docs/testing/manual-testing.md`.
- gRPC (work items): `backend/docs/postman/grpc-workitem.md` — manual Postman gRPC setup +
  a ready `grpcurl` script.
- Proving DB writes for real (psql/mongosh commands): also in
  `backend/docs/testing/manual-testing.md`.

## Common issues

| Symptom | Cause | Fix |
|---|---|---|
| `DATABASE_URL is not set` on startup | no `.env` in `backend/`, or ran from wrong dir | `cp .env.example .env` inside `backend/` |
| `goose: no migrations found` / migrate-up does nothing | old Makefile pointed at wrong dir | already fixed — migrations live in `internal/adapter/db/postgres/migrations` |
| Postman gRPC gets `UNIMPLEMENTED`/weird errors, not a timeout | something else already listening on the port you pointed Postman at | `lsof -nP -iTCP:<port> -sTCP:LISTEN`, set `GRPC_PORT` to a free one, restart backend, repoint Postman |
| Postman gRPC gets `UNAVAILABLE` | nothing listening at all — backend not actually running | check your terminal running `go run ./cmd/api` is still up |
| `bind: address already in use` on `docker compose up` | port 27017/5434 taken by another local service | remap host port in `docker-compose.yml` |
