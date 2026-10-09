# Chura backend

Go API (Fiber, GORM on Postgres, MongoDB for Work Items). Run the full stack
from the repository root as described in `frontend/README.md`; this page covers
local authentication and Project membership.

## Local auth setup

Every `/api/v1` route needs a Keycloak bearer token, and Project membership
and Project Roles live in Keycloak groups and user attributes, so the backend
talks to Keycloak both to validate tokens and through the Admin API.

### Environment variables

| Variable | Purpose |
|---|---|
| `KC_JWKS_ENDPOINT` | JWKS URL used to verify access tokens |
| `KC_BASE_URL` | Keycloak base URL for the Admin API, e.g. `http://localhost:8080` |
| `KC_REALM` | Realm name, `chura` |
| `KC_BACKEND_CLIENT_ID` | Confidential service-account client, `chura-backend` |
| `KC_BACKEND_CLIENT_SECRET` | Its secret (local dev value: `chura-backend-dev-secret`) |
| `GRPC_PORT` | Optional, default `9000`; change it if another local service holds 9000 |

`backend/.env.example` contains working local development values. Copy it to
`backend/.env` and adjust. The four `KC_*` variables above are required; the
backend refuses to start without them.

### Recreate local Keycloak after the realm change

The realm import adds the `chura-backend` client, email addresses, the user
`pim` and the user profile setting that lets only admins write Project Roles.
Keycloak only imports a realm into an empty database, so recreate it:

```bash
docker compose -f deploy/docker-compose.identity.yml down -v && docker compose -f deploy/docker-compose.identity.yml up -d
```

Warning: `down -v` deletes all local Keycloak data (users you created by hand,
sessions, settings).

### Seed users (development only)

All passwords are `test1234`.

| Username | Email | Chura role |
|---|---|---|
| `tonnam` | `tonnam@example.com` | Team Member |
| `pim` | `pim@example.com` | Team Member |
| `jojo` | `jojo@example.com` | Auditor |

### Calling the API with curl

```bash
TOKEN=$(curl -s -d client_id=chura-auth-client -d grant_type=password \
  -d username=tonnam -d password=test1234 \
  http://localhost:8080/realms/chura/protocol/openid-connect/token | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" http://localhost:<PORT>/api/v1/templates
```

Keycloak listens on 8080 locally, so run the backend on a different `PORT`
(for example 8083, as in `deploy/docker-compose.yml`).

Add a member by email (the person must be a Team Member):

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"email":"pim@example.com"}' \
  http://localhost:<PORT>/api/v1/projects/<projectId>/members
```
