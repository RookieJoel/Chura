# C1 — System context

Chura helps student teams run Agile projects (backlog, sprints, work items, reviews) while auditors inspect progress. Authentication and project membership are delegated to Keycloak; work-item notifications use email.

```plantuml
@startuml
!include <C4/C4_Context>

title System context — Chura

Person(team_member, "Team member", "Creates projects, manages backlog and sprints, and collaborates on work items.")
Person(auditor, "Auditor", "Reviews project progress, metrics, and sprint outcomes.")

System(chura, "Chura", "Web-based agile project management with templates, backlog, sprints, work items, analytics, and sprint review exports.")

System_Ext(keycloak, "Keycloak", "OIDC identity provider: login, JWT access tokens, realm roles, and project membership.")
System_Ext(smtp, "Email (SMTP)", "Delivers work-item notification emails to reporters and assignees.")

Rel_R(team_member, chura, "Manages projects, backlog, and sprints", "HTTPS")
Rel_R(auditor, chura, "Reviews dashboards and sprint summaries", "HTTPS")
Rel_R(chura, keycloak, "Authenticates users and resolves membership / roles", "OIDC / Admin API")
Rel_R(chura, smtp, "Sends notification emails", "SMTP")

SHOW_LEGEND()
@enduml
```

## Notes

- End users interact only with the Chura UI and API; they do not log into Postgres, MongoDB, or RabbitMQ directly.
- Keycloak runs as a separate Compose stack (`deploy/docker-compose.identity.yml`) but is a hard dependency for every protected API call.
- SMTP is optional in local dev but required for the RabbitMQ notification consumer in integrated environments (`backend/README.md`).
