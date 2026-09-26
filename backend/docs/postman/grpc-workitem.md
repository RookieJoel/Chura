# WorkItem gRPC test set

Postman **cannot import a gRPC test set from collection JSON** — Postman's own community forum
confirms non-HTTP (gRPC) collections have no export/import path yet
([source](https://community.postman.com/t/how-can-i-export-the-collection-with-grpc-requests/66848)).
So the gRPC part is covered two ways: a manual Postman setup (5 min, one-time) and a `grpcurl`
script (copy-paste, scriptable, CI-friendly). Both hit the same server on `localhost:$GRPC_PORT`
(default `9000`, see below).

Service has no reflection registered (`main.go`), so both routes need the `.proto` file directly:
`internal/adapter/handler/grpc/proto/workitem.proto`.

**Before anything else**, check port 9000 is actually free on your machine —
`lsof -nP -iTCP:9000 -sTCP:LISTEN`. If something else owns it (Docker containers from other
projects are the usual culprit), Postman/grpcurl will connect fine and get real-looking gRPC
errors (e.g. `UNIMPLEMENTED`) from the *wrong* server, which looks exactly like a Chura bug but
isn't. Fix: set `GRPC_PORT=<free-port>` in `backend/.env` (e.g. `19000`), restart
`go run ./cmd/api`, and use that port everywhere below instead of 9000.

## Option A — Postman manual gRPC setup

1. New request → **gRPC**.
2. Server URL: `localhost:9000` (or your `GRPC_PORT` value) — no `grpc://` prefix, Postman adds
   it. Make sure the field actually has your value in it, not the `<value>` placeholder text.
3. Service definition → **Import a .proto file** → select
   `backend/internal/adapter/handler/grpc/proto/workitem.proto`.
4. Pick a method from the dropdown (e.g. `ListWorkItems`), click **Use Example Message**, edit the
   JSON body, hit **Invoke**.
5. Repeat per method below. Save each as a request in a new collection so it's reusable.

## Option B — grpcurl (recommended, run from `backend/`)

Install once: `brew install grpcurl` (or `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`).

```bash
PROTO=internal/adapter/handler/grpc/proto/workitem.proto
ADDR=localhost:9000   # match your GRPC_PORT if you changed it

# 1. List services (sanity check the server is up)
# -import-path . is required — without it grpcurl errors with
# "must specify at least one import path"
grpcurl -plaintext -import-path . -proto "$PROTO" "$ADDR" list

# 2. Create a work item (client-streaming: one message, then close)
grpcurl -plaintext -import-path . -proto "$PROTO" -d '{
  "work_item": {
    "project_id": "proj-1",
    "title": "Wire sprint burndown chart",
    "description": "Aggregate story points per day",
    "type": "TASK",
    "status": "TO_DO",
    "priority": "HIGH",
    "assignee_id": "user-42",
    "reporter_id": "user-1",
    "story_points": 5
  }
}' "$ADDR" chura.workitem.v1.WorkItemService/CreateWorkItems
# copy the "id" from the response for the next steps

ID=<paste-id-here>

# 3. Get it back
grpcurl -plaintext -import-path . -proto "$PROTO" -d "{\"id\": \"$ID\"}" \
  "$ADDR" chura.workitem.v1.WorkItemService/GetWorkItem

# 4. List by project
grpcurl -plaintext -import-path . -proto "$PROTO" -d '{"project_id": "proj-1"}' \
  "$ADDR" chura.workitem.v1.WorkItemService/ListWorkItems

# 5. Update it (status -> IN_PROGRESS)
grpcurl -plaintext -import-path . -proto "$PROTO" -d "{
  \"id\": \"$ID\",
  \"work_item\": {
    \"project_id\": \"proj-1\",
    \"title\": \"Wire sprint burndown chart\",
    \"description\": \"Aggregate story points per day\",
    \"type\": \"TASK\",
    \"status\": \"IN_PROGRESS\",
    \"priority\": \"HIGH\",
    \"assignee_id\": \"user-42\",
    \"reporter_id\": \"user-1\",
    \"story_points\": 5
  }
}" "$ADDR" chura.workitem.v1.WorkItemService/UpdateWorkItem

# 6. Delete it
grpcurl -plaintext -import-path . -proto "$PROTO" -d "{\"id\": \"$ID\"}" \
  "$ADDR" chura.workitem.v1.WorkItemService/DeleteWorkItem

# 7. Confirm delete (expect NotFound, grpcurl exits non-zero — that's correct)
grpcurl -plaintext -import-path . -proto "$PROTO" -d "{\"id\": \"$ID\"}" \
  "$ADDR" chura.workitem.v1.WorkItemService/GetWorkItem
```

Verified live end-to-end against a real Postgres + Mongo (create → row/document appears,
update → mutates in place, delete → gone, get-after-delete → `NotFound`).

`type` / `status` / `priority` are proto enums — send the exact string names
(`TASK`, `USER_STORY`, `BUG` / `TO_DO`, `IN_PROGRESS`, `REVIEW`, `DONE`, `BLOCKED` /
`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`), not the domain lowercase values.
