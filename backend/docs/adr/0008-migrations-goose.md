**ADR-008: Schema Migrations**

| Title | Schema Migrations |
| :---- | :---- |
| **Context** | Postgres (ADR-006) schema needs versioned, reviewable migrations that also serve as sqlc's schema source (ADR-007). Candidates: golang-migrate, goose, atlas. |
| **Decision** | Use goose, with plain SQL migration files under `backend/db/migrations`, run via `make -C backend migrate-up` / `migrate-down` against `$DATABASE_URL`. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Simple CLI, plain SQL (no DSL to learn).<br>• Migration files double as the schema sqlc reads.<br>**Negative**<br>• `goose` binary isn't vendored — must be installed locally (`go install github.com/pressly/goose/v3/cmd/goose@latest`) or run via `go run` before `migrate-up`/`migrate-down` work. |
