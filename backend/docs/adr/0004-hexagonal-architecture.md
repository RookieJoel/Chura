**ADR-004: Backend Architecture Pattern**

| Title | Backend Architecture Pattern |
| :---- | :---- |
| **Context** | Chura's backend implements multiple business services (Project Configuration, Agile Execution, Analytics & Reporting — see `service.md`) with cross-service collaborators and server-side authorization (Team Member vs Auditor), inside a Go modular monolith (ADR-002). The team wants business logic (`service.md`'s operation catalog) decoupled from the HTTP framework and the database so either can change independently, and use cases stay testable without a live database. |
| **Decision** | Structure `backend/internal` as hexagonal ports and adapters: `domain` (entities, no framework deps), `port/in` (use-case interfaces), `port/out` (repository/gateway interfaces), `service` (use-case implementations — the only place business logic lives), `adapter/http` (driving adapter), `adapter/postgres` (driven adapter). |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Business logic testable via `port/in` against in-memory `port/out` fakes, no DB required.<br>• The HTTP framework and the database are both swappable behind their adapter boundary.<br>• Matches the Modular Monolith direction from ADR-002 with clear package boundaries.<br>**Negative**<br>• More files/indirection per feature than a typical MVC layout.<br>• The team must keep discipline about which layer a given piece of logic belongs in (see `.claude/skills/develop-feature-chura/SKILL.md` for the per-feature checklist). |
