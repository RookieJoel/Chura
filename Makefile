# Chura — local development entry point. Run `make` for the list of targets.
#
#   make setup   one-time (and re-runnable) dev setup: env files, Keycloak,
#                Postgres, MongoDB, migrations, frontend dependencies
#   make dev     run backend (go run) + frontend (pnpm dev) together

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

IDENTITY := docker compose -f deploy/docker-compose.identity.yml
# deploy/docker-compose.yml refuses to start without KC_BACKEND_CLIENT_SECRET,
# so pass the value from backend/.env (read at run time; setup creates it).
KC_SECRET = $$(sed -n 's/^KC_BACKEND_CLIENT_SECRET=//p' backend/.env 2>/dev/null)
# Postgres is published on the port DATABASE_URL in backend/.env points at.
PG_PORT = $$(sed -n 's|^DATABASE_URL=.*@[^:/]*:\([0-9]*\)/.*|\1|p' backend/.env 2>/dev/null)
DEPLOY = KC_BACKEND_CLIENT_SECRET="$(KC_SECRET)" CHURA_PG_PORT="$(PG_PORT)" docker compose -f deploy/docker-compose.yml

KC_URL := http://localhost:8080
API_PORT := 8083
DEV_PORTS := 8080 27017 3000

.PHONY: help setup check-tools check-ports env env-sync infra migrate frontend-deps verify-keycloak \
        dev backend frontend up down status logs reset-keycloak clean test lint

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## ---- one-time setup -------------------------------------------------------

setup: check-tools check-ports env infra migrate frontend-deps verify-keycloak ## One command dev setup (safe to re-run)
	@echo
	@echo "✅ Setup done."
	@echo "   make dev        → backend http://localhost:$(API_PORT) + frontend http://localhost:3000"
	@echo "   Keycloak admin  → $(KC_URL)/admin (admin / admin), realm 'chura'"
	@echo "   Test users      → tonnam (member), pim (member), jojo (auditor) — password test1234"

check-tools: ## Verify required tools are installed
	@missing=0; for t in docker go node pnpm openssl curl; do \
	  command -v $$t >/dev/null || { echo "❌ missing: $$t"; missing=1; }; done; \
	docker compose version >/dev/null 2>&1 || { echo "❌ missing: docker compose v2"; missing=1; }; \
	docker info >/dev/null 2>&1 || { echo "❌ Docker daemon is not running"; missing=1; }; \
	[ $$missing -eq 0 ] && echo "✔ tools: docker, go, node, pnpm, openssl, curl"; exit $$missing

check-ports: backend/.env ## Warn about dev ports held by something other than Chura's containers
	@for p in $(DEV_PORTS) "$(PG_PORT)" "$$(sed -n 's/^PORT=//p' backend/.env)" "$$(sed -n 's/^GRPC_PORT=//p' backend/.env)"; do \
	  nc -z 127.0.0.1 $$p 2>/dev/null || continue; \
	  name=$$(docker ps --filter "publish=$$p" --format '{{.Names}}' | head -1); \
	  [[ "$$name" == chura-* ]] && continue; \
	  [ -z "$$name" ] && name=$$(lsof -nP -iTCP:$$p -sTCP:LISTEN 2>/dev/null | awk 'NR==2 {print $$1}'); \
	  echo "⚠️  port $$p is already in use by '$${name:-a process of another user, e.g. a native Postgres}' — see SETUP.md › Common issues"; \
	done; echo "✔ port check done"

env: backend/.env frontend/.env.local env-sync ## Create env files from defaults (never overwrites values)

# Only created when missing; existing files are never overwritten.
backend/.env:
	cp backend/.env.example $@
	@echo "✔ created $@ from backend/.env.example"

frontend/.env.local:
	@{ echo "# Created by 'make setup' — local development only"; \
	  echo "AUTH_SECRET=$$(openssl rand -base64 32)"; \
	  echo "KEYCLOAK_ISSUER=$(KC_URL)/realms/chura"; \
	  echo "KEYCLOAK_CLIENT_ID=chura-auth-client"; \
	  echo "KEYCLOAK_CLIENT_SECRET="; \
	  echo "CHURA_API_URL=http://localhost:$(API_PORT)"; \
	  echo "NEXT_PUBLIC_WORK_ITEM_RPC_URL=ws://localhost:$(API_PORT)/ws/work-items"; \
	} > $@
	@echo "✔ created $@"

env-sync: backend/.env ## Append keys missing from an older backend/.env; warn on clashing values
	@added=""; while IFS= read -r line; do \
	  key=$${line%%=*}; [[ "$$line" =~ ^[A-Z_]+= ]] || continue; \
	  grep -q "^$$key=" backend/.env || { printf '%s\n' "$$line" >> backend/.env; added="$$added $$key"; }; \
	done < backend/.env.example; \
	[ -n "$$added" ] && echo "✔ added to backend/.env:$$added" || true
	@if grep -q '^PORT=8080$$' backend/.env; then \
	  sed -i.bak 's/^PORT=8080$$/PORT=$(API_PORT)/' backend/.env; \
	  echo "ℹ️  backend/.env PORT=8080 clashed with Keycloak → changed to $(API_PORT) (old file: backend/.env.bak)"; fi
	@grep -q '^DATABASE_URL=.*@localhost:[0-9]*/' backend/.env || \
	  echo "⚠️  backend/.env DATABASE_URL should look like $$(sed -n 's/^DATABASE_URL=//p' backend/.env.example)"

infra: backend/.env ## Start Keycloak, Postgres and MongoDB and wait until healthy
	@echo "▶ starting Keycloak (first start can take ~1 min)…"
	$(IDENTITY) up -d --wait
	@echo "▶ starting Postgres + MongoDB…"
	$(DEPLOY) up -d --wait postgre-db mongo-db

migrate: backend/.env ## Apply Postgres migrations (goose, in Docker)
	$(DEPLOY) run --rm --build agile-execution-migrations

frontend-deps: ## Install frontend dependencies
	pnpm --dir frontend install --frozen-lockfile

verify-keycloak: ## Check the local realm has the current config (backend client)
	@if curl -sf -o /dev/null -d client_id=chura-backend -d grant_type=client_credentials \
	    -d "client_secret=$(KC_SECRET)" $(KC_URL)/realms/chura/protocol/openid-connect/token; then \
	  echo "✔ Keycloak realm is up to date (chura-backend service account works)"; \
	else \
	  echo "❌ Keycloak realm is outdated or not ready: the chura-backend client is missing."; \
	  echo "   Your local Keycloak was created before the realm change. Run: make reset-keycloak"; \
	  exit 1; fi

## ---- day-to-day ------------------------------------------------------------

dev: ## Run backend + frontend locally (Ctrl-C stops both)
	@trap 'kill 0' EXIT INT TERM; \
	(cd backend && go run ./cmd/api) & \
	pnpm --dir frontend dev & \
	wait

backend: ## Run only the backend locally (reads backend/.env)
	cd backend && go run ./cmd/api

frontend: ## Run only the frontend (pnpm dev)
	pnpm --dir frontend dev

up: backend/.env ## Alternative: run Keycloak + full backend stack in Docker (API on :8083)
	$(IDENTITY) up -d --wait
	$(DEPLOY) up -d --build --wait

down: ## Stop all containers (data is kept)
	-$(DEPLOY) down
	-$(IDENTITY) down

status: ## Show container status
	@$(IDENTITY) ps; echo; $(DEPLOY) ps

logs: ## Follow container logs (Ctrl-C to stop)
	$(DEPLOY) logs -f & $(IDENTITY) logs -f; wait

test: ## Run backend (race) and frontend tests
	$(MAKE) -C backend test
	pnpm --dir frontend test

lint: ## Run backend vet + frontend lint/typecheck
	$(MAKE) -C backend lint
	pnpm --dir frontend lint
	pnpm --dir frontend typecheck

## ---- destructive (asks first) -----------------------------------------------

reset-keycloak: ## DELETE local Keycloak data and re-import the realm
	@read -r -p "This deletes ALL local Keycloak data (users, groups, Project memberships). Type 'yes': " a; [ "$$a" = yes ]
	$(IDENTITY) down -v
	$(IDENTITY) up -d --wait
	@$(MAKE) --no-print-directory verify-keycloak

clean: ## DELETE all local containers and volumes (Keycloak, Postgres, MongoDB)
	@read -r -p "This deletes ALL local Chura data (Keycloak, Postgres, MongoDB). Type 'yes': " a; [ "$$a" = yes ]
	-$(DEPLOY) down -v
	-$(IDENTITY) down -v
