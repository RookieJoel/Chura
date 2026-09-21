**ADR-006: Database**

| Title | Database |
| :---- | :---- |
| **Context** | `proposal.md`'s Non-functional/Data Integrity requirements demand valid relationships between Project, Ticket, Sprint, and Team Member, and consistent updates to assignments, Story Points, workflow status, and Sprint membership — relational integrity constraints, not a document store's strengths. |
| **Decision** | Use PostgreSQL as Chura's primary database. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Foreign keys and constraints enforce the relational-integrity requirements at the DB layer.<br>• Mature tooling (`sqlc`, `goose`) fits it directly.<br>• Well understood by student developers.<br>**Negative**<br>• Requires running/hosting a Postgres instance for local dev and deployment (vs. an embedded/file DB). |
