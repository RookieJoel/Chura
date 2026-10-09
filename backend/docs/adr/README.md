# Architecture Decision Records

The single ADR log for Chura, project-level and backend-specific alike —
numbered sequentially, never renumbered or deleted (a superseded decision
gets a new ADR that says so, and the old one's Status is updated to
`Superseded by NNNN`). 0001–0003 are transcribed from `proposal.md`
verbatim (that file remains the narrative source for the class
deliverable); 0004 onward were made during backend scaffolding.

A decision earns an ADR only when it's hard to reverse, surprising without
context, and the result of a real trade-off — not every config choice.

## Template

Same table format as `proposal.md`'s ADRs:

```markdown
**ADR-NNNN: <Title>**

| Title | <Title> |
| :---- | :---- |
| **Context** | What forces are at play — constraints, requirements, prior state. |
| **Decision** | What was decided, stated as an instruction ("Use X for Y"). |
| **Status** | Proposed \| Accepted \| Superseded by NNNN |
| **Consequences** | **Positive**<br>• ...<br>**Negative**<br>• ... |
```

## Index

- [0001](0001-frontend-framework.md) — Next.js + React
- [0002](0002-backend-framework.md) — Go (Golang)
- [0003](0003-css-ui-framework.md) — Tailwind CSS
- [0004](0004-workitem-api-architecture.md) — WorkItem API structure
- [0005](0005-hexagonal-architecture.md) — Backend architecture pattern
- [0006](0006-identity-provider-keycloak.md) — Identity provider (amended 2026-10-09: Project membership and Project Role live in Keycloak)
