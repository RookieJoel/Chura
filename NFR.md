# Non-functional requirements (Chura)

This document tracks **non-functional requirements (NFRs)** for Chura and how they are
verified in this repository. Functional requirements live in `proposal.md`; architecture
decisions live in `backend/docs/adr/`.

## Operational

| ID | Requirement | Notes / verification |
| --- | --- | --- |
| NFR-O1 | Web UI works in modern browsers | Next.js app at `http://localhost:3000` (`make dev`). |
| NFR-O2 | Responsive layout (desktop and mobile) | Tailwind-based UI; manual check on narrow viewports. |
| NFR-O3 | Data persists across sessions | Postgres (projects, sprints), MongoDB (work items); Docker volumes in `deploy/`. |
| NFR-O4 | Backup and recovery | Out of pilot scope; production runbooks TBD. |
| NFR-O5 | Consistent terminology and navigation | Product copy; manual UX review. |

## Performance

| ID | Requirement | Notes / verification |
| --- | --- | --- |
| NFR-P1 | Responsive common backlog/sprint/ticket actions | Target: interactive UI; no formal SLA in pilot. |
| NFR-P2 | Boards and dashboards load without noticeable delay | Manual check; k6 smoke in `backend/docs/adr/0007-load-testing-k6.md`. |
| NFR-P3 | Sprint metrics without blocking ticket edits | Analytics service async paths; integration tests. |
| NFR-P4 | Sprint review summaries in acceptable time | Analytics HTTP handlers; manual/k6 as needed. |
| NFR-P5 | Avoid redundant metric recomputation | Service-layer caching TBD per feature. |

## Security

| ID | Requirement | Implementation |
| --- | --- | --- |
| NFR-S1 | Authentication required for protected APIs | Keycloak OIDC; JWT validated on `/api/v1/*` (project, sprint, analytics). |
| NFR-S2 | Authorization by **Team Member** (`member`) and **Auditor** (`auditor`) realm roles | `backend/internal/adapter/middleware/auth.go` maps exactly one realm role to `domain.Actor`. |
| NFR-S3 | Team Members cannot access unrelated projects | Project membership via Keycloak groups; `RetrieveProjectData` returns 403 for non-members. |
| NFR-S4 | Auditor capabilities only for users with `auditor` role | Auditors are forbidden from mutating project boards (e.g. `POST /projects`). |
| NFR-S5 | Server-side enforcement (not UI-only) | All `/api/v1` routes on domain services use `KeycloakAuth` middleware. |
| NFR-S6 | Protect project, user, review, and auth data | TLS in production; secrets via env; no tokens in logs. |

**Out of scope for REST auth (pilot):** Work Item WebSocket (`/ws/work-items`) is not
JWT-protected today; treat it as a dev/integration surface behind Kong. Hardening is a
future NFR.

## Scalability

| ID | Requirement | Notes |
| --- | --- | --- |
| NFR-SC1 | More projects without core redesign | Modular monolith; Postgres per domain. |
| NFR-SC2 | More users (members and auditors) | Keycloak realm; horizontal scale of stateless API replicas via Consul/Kong. |
| NFR-SC3 | Growth in tickets, sprints, reviews | MongoDB work items; sprint/analytics tables. |
| NFR-SC4 | SE plugins without rewriting core | Template-driven project configuration. |
| NFR-SC5 | Optional future service split | Hexagonal boundaries per `backend/docs/adr/0005-hexagonal-architecture.md`. |

## Data integrity

| ID | Requirement | Notes |
| --- | --- | --- |
| NFR-D1 | No invalid/incomplete domain state | Validation in services + DB constraints (goose migrations). |
| NFR-D2 | Valid project–member–ticket–sprint relationships | FKs (e.g. sprints → projects); membership in Keycloak. |
| NFR-D3 | Tickets not assigned outside project membership | Enforced when assignment flows use project gateway (evolving). |
| NFR-D4 | Preserve status and sprint history | Domain events / persistence per feature. |
| NFR-D5 | Metrics derived from stored ticket/sprint data | Analytics service reads authoritative stores. |
| NFR-D6 | Consistent updates to assignments, points, status, sprint membership | Transaction boundaries in adapters. |
| NFR-D7 | Auditor reviews with author and timestamp | Reflection/review schemas in analytics migrations. |

## Local dev defaults

| Item | Value |
| --- | --- |
| API gateway | `http://localhost:8083` (`make up`) |
| Keycloak token endpoint | `http://localhost:8080/realms/chura/protocol/openid-connect/token` |
| OIDC client (password grant, dev only) | `chura-auth-client` |
| Seeded dev project id | `11111111-1111-4111-8111-111111111111` (`make seed-dev` → `backend/scripts/seed/dev_project.sql`) |
| Test users (password `test1234`) | `tonnam`, `pim` → **member**; `jojo` → **auditor** |

---

## Verifying API authorization (curl)

Prerequisites: Keycloak and the API stack are up (`make setup` then `make up`, or
`make dev`). Use the **gateway** base URL unless you are hitting a split port directly.

```bash
export KEYCLOAK=http://localhost:8080
export API=http://localhost:8083
```

### 1. Obtain access tokens (Resource Owner Password — **local dev only**)

**Team Member** (`tonnam`):

```bash
MEMBER_TOKEN=$(curl -sS -X POST "$KEYCLOAK/realms/chura/protocol/openid-connect/token" \
  -d client_id=chura-auth-client \
  -d grant_type=password \
  -d username=tonnam \
  -d password=test1234 \
  | jq -r .access_token)
echo "${MEMBER_TOKEN:0:20}..."
```

**Auditor** (`jojo`):

```bash
AUDITOR_TOKEN=$(curl -sS -X POST "$KEYCLOAK/realms/chura/protocol/openid-connect/token" \
  -d client_id=chura-auth-client \
  -d grant_type=password \
  -d username=jojo \
  -d password=test1234 \
  | jq -r .access_token)
```

If `access_token` is `null`, check Keycloak is healthy and the realm was imported
(`make reset-keycloak` re-imports from `keycloak/realm-config.json`).

### 2. Unauthenticated request → **401**

Protected routes must reject missing credentials:

```bash
curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" "$API/api/v1/whoami"
cat /tmp/chura-body.txt
# Expect: 401 and {"error":"unauthenticated"} (or equivalent)
```

### 3. Invalid token → **401**

```bash
curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -H "Authorization: Bearer not-a-real-jwt" \
  "$API/api/v1/whoami"
cat /tmp/chura-body.txt
# Expect: 401
```

### 4. Authenticated identity → **200** (`whoami`)

Confirms JWT signature, expiry, and realm roles reach the project service:

```bash
curl -sS -H "Authorization: Bearer $MEMBER_TOKEN" "$API/api/v1/whoami" | jq .
# Expect: authenticated true, user_id, email, roles including "member"

curl -sS -H "Authorization: Bearer $AUDITOR_TOKEN" "$API/api/v1/whoami" | jq .
# Expect: roles including "auditor" (not "member")
```

### 5. Read-only API access (both roles) → **200**

Templates are available to authenticated Team Members and Auditors:

```bash
curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -H "Authorization: Bearer $MEMBER_TOKEN" \
  "$API/api/v1/templates"
head -c 200 /tmp/chura-body.txt; echo
# Expect: 200 and a JSON array of templates

curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -H "Authorization: Bearer $AUDITOR_TOKEN" \
  "$API/api/v1/templates"
# Expect: 200
```

### 6. Team Member mutation → **201** (`POST /projects`)

Creating a project board requires the **member** realm role and a working Keycloak
Admin API (`chura-backend` service account):

```bash
curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -X POST \
  -H "Authorization: Bearer $MEMBER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Curl auth test","description":"NFR verification","template_id":"se"}' \
  "$API/api/v1/projects"
cat /tmp/chura-body.txt | jq .
# Expect: 201 and a project object with "id"
```

### 7. Auditor mutation denied → **403** (`POST /projects`)

Auditors authenticate but cannot create project boards (NFR-S4):

```bash
curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -X POST \
  -H "Authorization: Bearer $AUDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Should fail","template_id":"se"}' \
  "$API/api/v1/projects"
cat /tmp/chura-body.txt
# Expect: 403 and {"error":"forbidden"} (or wrapped service message)
```

### 8. Project-scoped access (membership)

`GET /api/v1/projects/{id}` requires the caller to be a **member of that project**
(Keycloak group), not merely a Team Member in the realm. The Postgres seed project
(`11111111-1111-4111-8111-111111111111`, from `make seed-dev`) does not add Keycloak membership by itself;
expect **403** `you are not a member of this project` until the user is added via
`POST /api/v1/projects/{id}/members` or project creation flow.

```bash
export PROJECT_ID=11111111-1111-4111-8111-111111111111

curl -sS -o /tmp/chura-body.txt -w "%{http_code}\n" \
  -H "Authorization: Bearer $MEMBER_TOKEN" \
  "$API/api/v1/projects/$PROJECT_ID"
cat /tmp/chura-body.txt
# Often 403 until project membership exists in Keycloak for this user
```

### 9. Public health (no bearer)

```bash
curl -sS -w "\n%{http_code}\n" "$API/health"
# Expect: 200 (routed to project service health)
```

### Expected status summary

| Check | Token | Endpoint | Expected HTTP |
| --- | --- | --- | --- |
| No auth | — | `GET /api/v1/whoami` | 401 |
| Bad JWT | invalid | `GET /api/v1/whoami` | 401 |
| Member identity | member | `GET /api/v1/whoami` | 200 |
| Auditor identity | auditor | `GET /api/v1/whoami` | 200 |
| List templates | member or auditor | `GET /api/v1/templates` | 200 |
| Create project | member | `POST /api/v1/projects` | 201 |
| Create project | auditor | `POST /api/v1/projects` | 403 |
| Project data | member, not in project | `GET /api/v1/projects/{id}` | 403 |
| Health | — | `GET /health` | 200 |

Automated coverage also exists in `backend/internal/adapter/handler/http/*_test.go` and
`backend/internal/adapter/middleware` tests; the steps above are the manual equivalent
for demos and NFR-S1–S5 sign-off.
