# Local development setup

Everything runs from the `Makefile` at the repo root. Run `make` to list targets.

## Prerequisites

- Docker Desktop (Compose v2), running
- Go, Node.js, pnpm
- `openssl`, `curl`

`make setup` checks for all of these and stops with a clear message if one is missing.

## Quick start

```bash
make setup   # one-time dev setup — safe to re-run
make dev     # backend + frontend together (Ctrl-C stops both)
```

| What | Where |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8083 (gRPC on 9000) |
| Keycloak admin | http://localhost:8080/admin — `admin` / `admin`, realm `chura` |
| Test users | `tonnam`, `pim` (members), `jojo` (auditor) — password `test1234` |

## What `make setup` does

1. **check-tools** — verifies the prerequisites and that the Docker daemon answers.
2. **check-ports** — warns when a dev port is held by something other than a `chura-*` container.
3. **env** — creates `backend/.env` (copied from `backend/.env.example`) and
   `frontend/.env.local` (with a fresh `AUTH_SECRET`) if they are missing.
   Existing files are never overwritten; keys added to `.env.example` later are appended.
4. **infra** — starts Keycloak, then Postgres and MongoDB, and waits until they are healthy.
5. **migrate** — applies Postgres migrations with goose (in Docker).
6. **frontend-deps** — `pnpm install --frozen-lockfile`.
7. **verify-keycloak** — checks the local realm has the `chura-backend` service account.

## Day-to-day targets

| Target | Purpose |
|---|---|
| `make dev` | Run backend (`go run`) and frontend (`pnpm dev`) |
| `make backend` / `make frontend` | Run only one side |
| `make up` | Alternative: run Keycloak + the full backend stack in Docker |
| `make down` | Stop all containers (data is kept) |
| `make status` | Show container status |
| `make logs` | Follow container logs |
| `make test` | Backend tests (with `-race`) + frontend tests |
| `make lint` | Backend vet + frontend lint and typecheck |

Destructive targets ask you to type `yes` first:

| Target | Deletes |
|---|---|
| `make reset-keycloak` | All local Keycloak data (users, groups, Project memberships), then re-imports the realm |
| `make clean` | All local containers and volumes — Keycloak, Postgres, MongoDB |

## Load balancing with Kong (`make up`)

`make up` runs 3 API replicas (`api-1`..`api-3`) behind the Kong API gateway.
Kong listens on :8083 (the API's usual port, so the frontend works unchanged)
and spreads requests round-robin across the replicas, dropping any that fail
`/health`. `make dev` is unchanged: one API, no Kong.

| What | Where |
|---|---|
| API through Kong | http://localhost:8083 |
| Kong Admin API | http://localhost:8001 |
| Konga dashboard (read-only) | http://localhost:1338 — `admin` / `admin1234` |

On Windows, run `make` from Git Bash (`winget install ezwinports.make` if it is missing).

**Config** lives in `deploy/kong/kong.yml` (Kong runs DB-less). Edit that file, then
`docker compose -f deploy/docker-compose.yml restart kong`. Konga only displays it;
creating or editing entities in Konga fails by design.

**See it work.** Every API response carries `X-Instance: api-N`, the replica that served it.
`/health` also puts it in the body, so refreshing http://localhost:8083/health in a
browser shows `"instance"` changing.

1. Konga → **Upstreams → chura-api-upstream → Targets**: the 3 replicas and their health.
2. Requests rotate across replicas (also visible in the browser's Network tab):
   ```bash
   for i in $(seq 1 6); do curl -si localhost:8083/health | grep -i x-instance; done
   ```
   Konga's **Dashboard** request counter goes up as you send them.
3. Failover: `docker stop chura-agile-execution-api-2`. Within ~10 s Kong marks it
   unhealthy (Konga, or `curl -s localhost:8001/upstreams/chura-api-upstream/health`)
   and only `api-1` / `api-3` answer. `docker start chura-agile-execution-api-2` brings it back.
4. Weights: set e.g. `200 / 100 / 100` in `kong.yml` and restart Kong; `api-1` then gets half the traffic.

In this mode the API's gRPC port (9000) is internal to each container; it is not published.

## Common issues

### A port is already in use

`check-ports` prints `⚠️ port N is already in use by '<name>'`. Dev ports:
8080 (Keycloak), 3000 (frontend), 27017 (MongoDB), 8083 / 9000 (API), and the
Postgres port from `DATABASE_URL` in `backend/.env` (default 5432).

- **Another Docker container** — stop it (`docker stop <name>`).
- **A native Postgres** (e.g. Homebrew, Postgres.app) on 5432 — either stop it, or
  change the port in `DATABASE_URL` in `backend/.env` (e.g. `localhost:5434`).
  The Makefile publishes the Chura Postgres container on whatever port `DATABASE_URL` uses.
- **A process of another user** — `lsof` cannot see it without `sudo`; check with
  `sudo lsof -nP -iTCP:<port> -sTCP:LISTEN`.

### `Keycloak realm is outdated or not ready`

Your Keycloak volume was created before the current `keycloak/realm-config.json`.
Keycloak imports the realm only on first start, so run `make reset-keycloak`
(this deletes local Keycloak data).

### `Docker daemon is not running`, or Docker keeps stopping by itself

- Make sure Docker Desktop is open and `docker info` answers.
- **Check free disk space** (`df -h /System/Volumes/Data` on macOS). When the host
  disk is full, containers die and Docker Desktop stops the engine; the daemon may
  hang until Docker Desktop is force-quit and reopened. Keep at least ~15 GB free.
- Reclaim space without losing data:
  ```bash
  docker system df              # what uses the space
  docker builder prune -af      # build cache
  docker image prune -a         # images no container uses
  ```
  Avoid `docker volume prune` and `docker system prune --volumes` — they delete
  the Postgres, MongoDB and Keycloak data.
- With **Resource Saver** enabled, Docker Desktop pauses the VM after a few idle
  minutes and shows "resuming" on the next command. That is expected, not a crash.

### `backend/.env` changed after `make setup`

`env-sync` may append missing keys, and it changes `PORT=8080` (which clashes with
Keycloak) to `PORT=8083`, keeping the previous file as `backend/.env.bak`.
Delete `.env.bak` once you have checked the change.
