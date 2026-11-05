# C2 — Containers

Local and Docker deployments expose a single HTTP entry on Kong (`:8083`). Domain services register with Consul; Kong resolves `*.service.consul` and routes by path (`deploy/kong/kong.yml`). The Next.js frontend talks to Kong and Keycloak; it proxies some API calls through Next.js route handlers.

```plantuml
@startuml
!include <C4/C4_Container>

LAYOUT_WITH_LEGEND()
title Container diagram — Chura

System_Ext(keycloak, "Keycloak", "OIDC identity provider: login, JWT access tokens, realm roles, and project membership.")

Person(team_member, "Team member", "Student project participant")
Person(auditor, "Auditor", "Reviewer with read-focused access")

System_Boundary(chura_boundary, "Chura platform") {
    Container(web_app, "Web application", "Next.js, React, NextAuth", "UI for boards, backlog, sprints, and reviews. Session via Keycloak; API calls to Kong.")
    Container(kong, "API gateway", "Kong 2.8 (DB-less)", "Public HTTP/WebSocket entry (:8083). Path-based routing to domain services via Consul DNS.")

    Container(project_svc, "Project service", "Go / Fiber", "Templates, projects, members, whoami, and health.")
    Container(analytics_svc, "Analytics service", "Go / Fiber", "Sprint summaries, metrics, reflections, and review exports.")

    Boundary(agile_execution, "Agile execution service") {
        Container(sprint_svc, "Sprint service", "Go / Fiber", "Sprint CRUD and lifecycle.")
        Container(workitem_svc, "Work item service", "Go (gRPC over websocket)", "Work items, realtime access, events, and notifications.")
    }

    ContainerDb(postgres, "Application database", "PostgreSQL", "Projects, sprints, and analytics data.")
    ContainerDb(mongo, "Work item store", "MongoDB", "Work item documents and backlog state.")
    ContainerQueue(rabbitmq, "Message broker", "RabbitMQ", "Notification queue.")

    Container(consul, "Service registry", "HashiCorp Consul", "Health checks and DNS for service upstreams.")
}

System_Ext(smtp, "Email (SMTP)", "Outbound mail for notifications.")

Rel_U(sprint_svc,keycloak,"JWKS verification and authorization","HTTP")
Rel_U(workitem_svc,keycloak,"JWKS verification and authorization","HTTP")
Rel_U(sprint_svc,keycloak,"JWKS verification and authorization","HTTP")
Rel_U(project_svc,keycloak,"JWKS verification and authorization","HTTP")

Rel_D(team_member, web_app, "Uses", "HTTPS")
Rel_D(auditor, web_app, "Uses", "HTTPS")
Rel_R(web_app, kong, "Calls APIs", "HTTP / WebSocket")

Rel_D(kong, consul, "Discovers services", "DNS")
Rel_D(project_svc, consul, "Registers", "HTTP")
Rel_D(sprint_svc, consul, "Registers", "HTTP")
Rel_D(workitem_svc, consul, "Registers", "HTTP")
Rel_D(analytics_svc, consul, "Registers", "HTTP")
Rel_D(project_svc, postgres, "Uses", "SQL")
Rel_D(sprint_svc, postgres, "Uses", "SQL")
Rel_D(analytics_svc, postgres, "Uses", "SQL")
Rel_D(workitem_svc, mongo, "Uses", "MongoDB")
Rel_D(workitem_svc, rabbitmq, "Publishes and consumes", "AMQP")
Rel_D(workitem_svc, smtp, "Sends notifications", "SMTP")

SHOW_LEGEND()
@enduml
```

Kong discovers backend services through Consul rather than depending on each
service as a separately configured route. Keycloak is intentionally omitted
from this container view for readability. It remains a cross-cutting
dependency: the web application uses OIDC, backend services validate JWTs
with its JWKS endpoint, and project-related operations use its Admin API for
membership and roles. Keycloak persists its realm data in a separate
PostgreSQL container (`deploy/docker-compose.identity.yml` →
`keycloak-db`).

## Routing reference (Kong → service)

| Path prefix / pattern                                                | Upstream (Consul name) | Port |
| -------------------------------------------------------------------- | ---------------------- | ---- |
| `/api/v1/templates`, `/api/v1/projects`, `/api/v1/whoami`, `/health` | `chura-project`        | 8059 |
| `/api/v1/sprints`                                                    | `chura-sprint`         | 8057 |
| Work-item HTTP + `/ws/work-items`                                    | `chura-workitem`       | 8058 |
| Sprint analytics, reflections, review exports                        | `chura-analytics`      | 8056 |

## Dev-only containers (omitted from diagram)

| Container                    | Role                                                |
| ---------------------------- | --------------------------------------------------- |
| Konga                        | Read-only Kong dashboard (`:1338`)                  |
| mongo-express                | MongoDB admin UI (`:8081`)                          |
| `agile-execution-migrations` | One-shot goose migrations before API services start |

## Alternate layout

`make backend-split` runs the four Go services on the host without Kong; the frontend still targets the same API base URL when Kong is on `:8083`.
