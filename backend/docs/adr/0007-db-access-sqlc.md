**ADR-007: Database Access Layer**

| Title | Database Access Layer |
| :---- | :---- |
| **Context** | `internal/adapter/postgres` (ADR-004) needs a way to talk to Postgres (ADR-006). Candidates: raw `database/sql`, an ORM (gorm/ent), or sqlc (codegen from hand-written SQL). `service.md` already documents the operation catalog in explicit, reviewable form; the team wants the same explicitness for queries rather than ORM-generated SQL. |
| **Decision** | Use sqlc, configured in `backend/sqlc.yaml`, generating into `backend/internal/adapter/postgres/sqlcgen` from `.sql` files in `backend/db/queries` against the schema in `backend/db/migrations`. `sqlcgen` is never hand-edited — always regenerated via `make -C backend generate`. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• SQL stays explicit and reviewable in PRs.<br>• Generated Go types give compile-time safety without ORM magic/reflection.<br>**Negative**<br>• Requires running `sqlc generate` after every query/schema change and committing the generated output.<br>• An extra toolchain step (`sqlc` binary) contributors must have installed. |
