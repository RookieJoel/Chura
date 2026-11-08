# Frontend wiring backlog

Dependency-ordered, demoable slices for the Chura frontend. Phase A
(project foundation) is complete. Phase B wires the rest of the product
onto that foundation so Board, Backlog, Sprints, and later analytics use
the **selected Project**, not a seed env id.

Sources: `proposal.md`, `service.md`. Items that need a product decision
are marked instead of being treated as agreed features.

## Phase A — Project foundation (done)

### In scope (Phase A)

- Project discovery, selection, templates, and onboarding.
- Project membership and roles.
- Project-level Auditor overview and feedback.
- Project settings and permanent deletion.
- Shared project navigation needed for these capabilities.

### Explicitly deferred from Phase A

Sprint creation, planning, execution, analytics, and reports; Work Item
creation, editing, assignment, comments, search, and filtering; project
views whose implementation would require those features. Permanent
Project deletion must remove associated Sprint and Work Item records,
but Phase A does not add or change Sprint or Work Item functionality.

Archive/restore, notification bell, and Reports nav were left **OUT**
(not shipped) rather than half-built.

---

## Ordered work slices

### 1. Project discovery contract — backend

**Blocked by:** None  
**Source:** Project Management requirements; current frontend calls
`GET /api/v1/projects`.

- [x] Implement `ListAccessibleProjects` with database-side scope filtering
  and pagination. Team Members see their memberships; Auditors see all
  Projects.
- [x] Add the authenticated project-list handler and register
  `GET /api/v1/projects`.
- [x] Define the paginated response contract
  (`items`, `page`, `page_size`, `total`), defaults (page 1, page size 15),
  and maximum page size (30).
- [x] Document the endpoint in `backend/docs/api/openapi.yaml`.

**Done:** The endpoint returns only the caller's authorized project
summaries. `go test ./...` and `go vet ./...` pass. Slice 2 updated the
frontend client to consume the paginated response.

### 2. Project discovery and selection — frontend

**Blocked by:** Slice 1's endpoint and response contract  
**Source:** Project selection implied by project creation/retrieval use cases.

- [x] Keep the existing Next.js proxy and typed project client aligned with
  the paginated backend contract; preserve query parameters and fetch all
  pages before returning results.
- [x] Mount `ProjectProvider` around the application shell.
- [x] Replace the hardcoded TopNav project name with the selected project's
  name and connect the selector to provider state.
- [x] Show loading, no-accessible-projects, and retryable error states.
- [x] Preserve a valid selection and replace stale or inaccessible stored
  selections with an accessible project.
- [x] Manually exercise the authenticated selector in a browser: load and
  select Projects, reload to verify selection persistence, and verify empty,
  discovery-error, and retry behavior.

**Done:** API, route-proxy, selector-state, selection-persistence, and
AppShell integration tests pass. Authenticated browser verification
confirmed live list, persistence, empty and error/retry states.

### 3. Project templates and creation — vertical slice

**Blocked by:** Slice 2 for selecting the created project; template discovery
can be implemented independently  
**Source:** Project Configuration use case and Project Management requirements.

- [x] Add typed frontend API clients for listing and retrieving templates;
  add same-origin authenticated proxy routes for both existing backend
  template endpoints and test query/response/error forwarding.
- [x] Add a same-origin authenticated create-project proxy/client that
  forwards the request and surfaces backend validation and authorization
  errors.
- [x] Add a Create Project entry to the project selector that remains
  available when the Team Member has no accessible Projects.
- [x] Build a project creation flow for project name, description, and
  explicit selection from the available templates.
- [x] Submit creation through the existing backend project-create route and
  surface validation and authorization errors; ensure the creator receives
  the template's creator role, which defines Project ownership.
- [x] Build the `/projects/:id` minimal overview route and, on successful
  creation, select the returned Project and navigate to it without Sprint or
  Work Item UI.
- [x] Add frontend tests for template loading/selection, successful creation,
  empty templates, validation/API failures, and authenticated proxy behavior.

**Done:** Create-to-overview flow works with authenticated proxies;
overview omits Sprint/Work Item controls by design for this slice.

### 4. Project members and roles — vertical slice

**Blocked by:** Slices 2 and 3 for selected-project context, the Project
overview, and template-defined role options  
**Source:** Project Configuration use case and User Access & Role Management
requirements.

- [x] Confirm the existing authenticated `GET /api/v1/projects/:id` response
  provides the member fields needed by the UI; it already includes project
  members.
- [x] Add same-origin authenticated proxy routes and typed clients for
  project detail, adding a member, assigning/updating a role, and removing a
  member; test authentication, request forwarding, and backend errors.
- [x] Use the selected Project's template definition to offer only its
  allowed roles.
- [x] Replace fixture-backed member data in the TopNav and sidebar with the
  selected project's members.
- [x] On the Project overview, allow only a current Project owner (a member
  holding the template's creator role) to add an existing Team Member by
  email, assign/update template-defined roles, or remove a member.
- [x] Add or update backend authorization and routes so membership changes
  enforce owner-only access; prevent removing or demoting the last owner.
- [x] Newly added members have no Project Role until an owner assigns a
  role allowed by that Project's template.
- [x] Test allowed and denied membership/role changes, including Auditor
  attempts, non-member attempts, invalid roles, and attempts to remove or
  demote the last owner.
- [x] Cover loading, empty, and error states without showing fixtures as live
  project data.

**Done:** Shell members match the selected project; only a Project owner
can manage membership and roles; unauthorized changes are rejected by the
backend.

### 5. Project-level Auditor review and feedback — vertical slice

**Blocked by:** Slices 1 and 3 for Auditor-visible discovery and the
Project overview. Slice 7 must later remove this slice's assignments and
feedback during permanent deletion.  
**Source:** Auditor Review & Feedback requirements in `proposal.md`.

- [x] Add Project Auditor self-assignment, allowing multiple Auditors per
  Project; an assigned Auditor cannot unassign until the Project is
  permanently deleted.
- [x] Attach feedback to the Project itself; exclude comments on Work Items
  and Sprint reviews.
- [x] Decide the feedback entry fields and validation limits before defining
  the persistence and API contract; do not assume a rating or fixed form.
- [x] Allow only assigned Auditors to submit feedback; allow Team Members
  belonging to the Project to read it, but not reply.
- [x] Add backend persistence that records feedback author and timestamp,
  preserves append-only feedback history, and stores Project Auditor
  assignments.
- [x] Add authenticated create/read API operations, response/error contracts,
  and authorization tests for assignment, reassignment prohibition,
  Auditor submission, and member read-only access. Repeated self-assignment
  must not create duplicate assignments.
- [x] Build an Auditor project overview limited to project-level information
  that does not depend on Sprint or Work Item functionality.
- [x] Build UI for Auditors to self-assign and submit feedback and for
  Project members to read feedback history; cover loading, empty, success,
  and error states.

**Done:** Feedback is required trimmed body text (1–4000 chars),
newest-first, append-only. Auditor self-assign and feedback live on
`/projects/:id`.

### 6. Project context consistency

**Blocked by:** Slices 2–5 and 7, so all project-scoped views and mutations
can share and verify the selected Project context  
**Source:** Project selection and membership use cases.

- [x] Ensure project-scoped data (overview, members, feedback, Auditor
  assignments, and settings) uses the selected Project rather than a
  development seed ID.
- [x] On Project switch, immediately clear the previous Project's data, show
  loading, then show the selected Project's data or a retryable error.
- [x] Prevent late responses for a previous selection from repopulating
  stale Project data; handle stale or inaccessible selections safely.
- [x] After deleting the selected Project, clear its persisted selection and
  select another accessible Project or show the empty state.

**Done:** Provider clears selection detail on switch; overview keys fetches
by route project id; delete clears persisted selection.

### 7. Scope decisions and shell cleanup

**Blocked by:** Slices 3–5 for Project creation, owner/member rules, and
feedback/assignment cleanup;
independent of the implementation slices for notification cleanup

- [x] Add authenticated backend update/delete operations and same-origin
  authenticated frontend proxies/typed clients for Project settings and
  deletion.
- [x] Add Project name and description editing, restricted to a Project owner
  (the member holding the template's creator role).
- [x] Allow template changes only when the Project has no Work Items or
  Sprints and only its creator is a member. Ensure that owner receives the
  new template's creator role; no other member roles need mapping under this
  eligibility rule. Enforce eligibility and ownership in the backend, not
  only in the UI.
- [x] Add permanent Project deletion for the Project owner. Delete all
  Project-owned data, including associated membership groups, Auditor
  assignments, feedback, Sprints, and Work Items; retain user accounts.
- [x] Define and test failure handling across database deletion and identity
  provider group deletion so partial cleanup is surfaced and can be safely
  retried; never report deletion success while Project-owned data remains.
- [x] Test unauthorized deletion/edits and template-change rejection for
  Projects with Work Items, Sprints, or additional members.
- [x] Require an explicit confirmation before permanent deletion and ensure
  the UI explains that deletion cannot be undone. Confirm the exact
  confirmation interaction before implementation.
- [x] Do not implement archive or restore. Permanently deleted Projects and
  their associated data are not available in the application.
- [x] Remove the notification bell until notification behavior and an API
  scope are defined.
- [x] Review remaining project-level navigation and controls so each
  navigates, performs an action, or is clearly disabled.

**Done:** Owner settings and permanent-delete Modal live on `/projects/:id`.
Notification bell removed; Reports and Settings nav remain commented out
(**OUT** until scoped).

---

## Phase B — Wire the rest of the product onto selected Project

Phase A left Board / Backlog / Sprints on a seed env id. Slices **8–9**
fixed selected-project wiring and overview discoverability. The sessions
below finish the product in **one chat session each**, backend work first
when an API is missing.

### Backend readiness (inventory)

| Area | Backend today | Gap |
|---|---|---|
| Sprint CRUD + status | **Exists** (`/api/v1/sprints`) | Optional: enforce “cannot start with zero Work Items” |
| Work Item CRUD (WS) | **Exists** (`/ws/work-items`) | — |
| WI validate vs template | **Exists** (`POST /projects/:id/work-items/validate`) | Frontend wired (Slice 12) |
| WI `blocked` status | **Exists** (status enum) | Reason stays in `features.blocked_reason` (optional first-class field deferred) |
| WI comments | **Exists** (`GET/POST /api/v1/work-items/:id/comments`) | Frontend wired (Slice 14) |
| WI backlog rank/order | **Missing** | Product decision; defer or add rank field |
| Work summary | **Exists** (`GET .../work-summary`) | Frontend client + Board UI (Slice 15) |
| Sprint metrics | **Exists** (`GET .../metrics`) | Frontend client + Board UI (Slice 15) |
| Reflections | **Exists** (`GET/PUT .../reflections`) | Frontend client + Reports UI (Slice 16) |
| Review summary + export | **Exists** (`.../review-summary`, `.../export`) | Frontend client + Reports UI (Slice 16) |
| Project-level Auditor feedback | **Exists** (Phase A) | — |
| Email notifications (mailing) | **Exists** (backend `NotificationService` + email sender; e.g. sprint finished / WI events via consumer) | Not a frontend slice — no in-app inbox API |
| In-app notification bell / feed | **Missing** | Keep **OUT** until scoped |
| Global search | **Missing** | Keep **OUT** until scoped |
| Archive / restore | **OUT** by product decision | — |

### Still OUT (do not build until scoped)

- Archive / restore Projects.
- In-app notification bell / feed (mailing email already exists; bell needs a separate list/read API + UX).
- Global Work Item search (needs query contract).

### Session map (do in order)

| Session | Slice | Focus | Est. |
|---|---|---|---|
| Done | 8–9 | Selected-project wiring + Overview nav | — |
| Done | 10 | Sprint lifecycle polish (mostly FE) | — |
| Done | 11 | Board polish (FE) | — |
| **3** | 12 | Backlog / WI planning + validate API (FE) | 1 session |
| Done | 13 | **Backend** WI comments (`features.blocked_reason` kept) | — |
| Done | 14 | **Frontend** WI comments + hide stub | — |
| Done | 15 | Analytics clients + metrics / work-summary UI | — |
| Done | 16 | Reflections + review summary + export + Reports nav | — |
| Done | 17 | Auditor observability (UC-04) on top of 8–16 | — |

Phase B execution + Auditor observability slices are complete.

---

### 8. Selected-project execution wiring — foundation blocker

**Session:** done (partial tests still open)  
**Blocked by:** Phase A slices 2 and 6

- [x] Pass `selectedProjectId` into sprint / work-item / members clients.
- [x] Board / Backlog / Sprints / WI detail reload on selection change.
- [x] Empty state when no Project is selected (`NoProjectSelected`).
- [ ] Automated tests for `project_id` plumbing; manual verify new Project
  shows empty sprints/WIs vs seed Project.

**Done when:** Selector changes Board/Backlog/Sprints data; seed env not
required for normal use.

### 9. Project overview discoverability and shell honesty

**Session:** done  
**Blocked by:** Slice 8

- [x] TopNav **Overview** + Sidebar project name / Overview → `/projects/:id`.
- [x] Remove fake search and backlog “View settings”.
- [x] Reports nav stays commented out (**OUT**).

**Done when:** Settings/members/feedback reachable without typing a URL.

### 10. Session 1 — Sprint lifecycle polish

**Blocked by:** Slice 8  
**Backend:** Sprint CRUD **exists**. Optional small backend rule below.  
**Source:** UC-03.

- [x] Confirm create/edit/start/end/delete already work on selected Project;
  fix remaining UX gaps (empty, errors, disable create without selection).
- [x] Show planned story-point total for a Sprint (sum of assigned WIs).
- [x] **Product decision:** unfinished items at Sprint end — document
  current backend behavior; no silent fake carry-over UI.
- [x] **Optional backend:** reject Start Sprint when the Sprint has zero
  Work Items (proposal alternate flow); add service test + OpenAPI note.
- [x] **Frontend:** if backend rule lands, surface the error; else warn in
  UI before start.
- [x] Tests + manual: full Sprint lifecycle on a non-seed Project.

**Done:** Empty-start is enforced in `SprintService` (create/update →
`active` with zero WIs → 400 on `status`). Completing a Sprint does **not**
move unfinished Work Items to the Backlog or another Sprint; they remain
assigned and appear as analytics carry-over only. Sprints page shows planned
points / WI counts, disables Start when empty, and confirms End without fake
carry-over UI.

### 11. Session 2 — Board polish

**Blocked by:** Slices 8 and 10  
**Backend:** WI WS + statuses **exist**.  
**Source:** UC-03 execution.  
**Session:** done

- [x] Confirm Board loads active Sprint WIs for selected Project only.
- [x] Status changes persist; column point totals correct.
- [x] Empty active-sprint state links to Sprints.
- [x] Document WS auth (`auth.<jwt>`, `127.0.0.1`) in README if not already.
- [x] **Defer explicitly:** drag-and-drop; live multi-user sync (keep
  request/response WS).
- [x] Tests + manual: Board on seed vs empty Project.

**Done:** `ProjectProvider` lives in root `Providers` so Board/Backlog/Sprints
can read selection outside `AppShell`. Board loads `getActiveSprint(projectId)`
then WIs for that sprint only; status updates go through request/response WS;
empty active-sprint state links to `/sprints`. DnD and live multi-user sync
remain deferred (documented in `frontend/README.md`).

### 12. Session 3 — Backlog / Work Item planning + validate

**Blocked by:** Slice 8; templates from Phase A  
**Backend:** WI WS + `POST /projects/:id/work-items/validate` **exist**.  
**Missing backend (defer to Session 4):** comments; optional
`blocked_reason` field.  
**Source:** UC-02.

- [x] Template-gated WI types on create/edit (from selected template).
- [x] Acceptance Criteria UI for User Stories (store via `features` or
  agreed field until a dedicated API exists).
- [x] Wire frontend to `POST /api/v1/projects/:id/work-items/validate`
  (proxy + client); show violations before/on save for SE mode.
- [x] Assignee / reporter / sprint dropdowns use selected Project (already
  partially done — finish gaps).
- [x] Add **Blocked** to board/detail workflow; require a reason in UI —
  persist as `features.blocked_reason` **or** land Session 4 field first.
- [x] Hide dead comment composer until Session 5.
- [x] **Defer:** backlog drag-rank (no backend rank) — remove any fake
  reorder control.
- [x] Tests + manual: create SE User Story with AC on a new Project.

**Done when:** Planning WIs on a non-seed Project uses template rules and
validation; no no-op comment UI.

### 13. Session 4 — Backend: Work Item comments (+ blocked reason)

**Blocked by:** None (API design); FE consumes in Session 5  
**Backend today:** **Exists** — REST comments on workitem service.  
**Source:** proposal ticket comments; Auditor ticket comments.  
**Session:** done

- [x] Design comment resource: `id`, `work_item_id`, `author_user_id`,
  `author_name`, `author_email`, `body`, `created_at` (append-only).
- [x] Persist comments in Mongo `work_item_comments` (alongside work items).
- [x] Authz: Project members can read/create; assigned Auditors can
  read/create; others 403.
- [x] REST under `/api/v1/work-items/:id/comments` (list newest-first +
  create); Kong routes to `chura-workitem`; OpenAPI updated.
- [x] **Optional same session:** deferred — keep `features.blocked_reason`
  from Session 3 (no first-class WI field yet).
- [x] Service + handler tests; OpenAPI updated.

**Done:** Members and assigned Auditors can `GET`/`POST`
`/api/v1/work-items/{id}/comments`. Comments are append-only, newest first,
body 1–4000 runes. Project delete cleans comments with Work Items.

### 14. Session 5 — Frontend: Work Item comments

**Blocked by:** Session 4  
**Source:** Detail page collaboration.
**Session:** done

- [x] Same-origin proxy + typed client for comment list/create.
- [x] Replace stub `listComments → []` and enable composer.
- [x] Show author + timestamp; newest-first order documented in UI.
- [x] Loading / empty / error states; unauthorized errors clear.
- [x] Tests + manual: member and Auditor comment on a WI.

**Done:** Comments persist across reload; stub empty list removed.

### 15. Session 6 — Analytics clients + metrics / work-summary UI

**Blocked by:** Slices 10–11 (need real Sprint/WI data)  
**Backend:** **Exists** — `GET /sprints/:id/work-summary`,
`GET /sprints/:id/metrics`.  
**Source:** UC-03 metrics; UC-04 observability inputs.

- [x] Frontend proxies (if needed) + typed clients for work-summary and
  metrics.
- [x] Sprint detail or Board header: show summary buckets (planned /
  completed / carry-over / blocked) and key metrics.
- [x] Missing data → partial UI, not fake charts.
- [x] Tests + manual: metrics match Sprint WIs for selected Project.

**Done when:** Team Member can see work-summary/metrics for a Sprint
without leaving the app.

### 16. Session 7 — Reflections + review summary + Reports nav

**Blocked by:** Session 6 (clients/patterns); Sprint exists  
**Backend:** **Exists** — reflections, review-summary, export.  
**Source:** UC-05.

- [x] Clients for `GET/PUT .../reflections`, `GET .../review-summary`,
  `GET .../review-summary/export`.
- [x] UI: submit/replace my reflection from template questions.
- [x] UI: generate review summary (metric query optional); preview;
  download/export.
- [x] Enable **Reports** sidebar item pointing at this surface (ends
  Reports **OUT** for nav).
- [x] Tests + manual: reflection + summary + export on a Sprint.

**Done when:** Sprint Review flow works end-to-end; Reports nav is real.

### 17. Session 8 — Auditor observability (UC-04)

**Blocked by:** Sessions 1–7 (execution + metrics + comments)  
**Backend:** Project feedback **exists** (Phase A); metrics/comments from
earlier sessions.  
**Source:** UC-04.

- [x] Auditor path: select any Project → Backlog / Board / Sprints /
  metrics usable (no seed dependency).
- [x] Ticket inspection + comments (Session 5) work for Auditors.
- [x] Keep Phase A project-level feedback on Overview unchanged.
- [x] If metrics missing, show available project info only.
- [ ] Manual: `jojo` reviews a non-seed Project end-to-end.

**Done when:** Auditor can inspect progress without seed env coupling.

**Done:** Pure Auditors get read-only Backlog / Board / Sprints / WI detail
on the selected Project (no `NEXT_PUBLIC_CHURA_PROJECT_ID` list fallback).
Comments stay available; 403s point Auditors to Overview self-assign.
Missing metrics / inactive Sprint still show available Project info.

---

## Cross-slice quality checks

### Phase A

- [x] Map each agreed Phase A requirement to at least one acceptance
  criterion and automated test.
- [x] Cover role boundaries, cross-project access, invalid input, empty
  data, and backend/proxy failures at the relevant API/UI seams.
- [x] Verify permanent deletion removes all Project-owned records and
  membership groups while retaining user accounts; verify failed cleanup
  does not appear successful.
- [x] Verify each completed Phase A slice end to end; do not treat a hidden
  UI control as authorization.
- [x] Update API documentation when backend contracts are added or changed.

### Phase B

- [x] No Board / Backlog / Sprints / members client path hard-requires
  `NEXT_PUBLIC_CHURA_PROJECT_ID` for normal authenticated use.
- [ ] Switching Projects never shows Project A’s sprints or work items as
  Project B’s.
- [ ] Template-gated WI types and SE validation match the selected Project’s
  template.
- [ ] OUT items (archive, notification bell, global search) stay absent until
  scoped; Reports nav appears only after Session 7.
- [x] Update OpenAPI when WI comments land (Session 4); `blocked_reason` stays in features.
- [x] Analytics FE uses existing work-summary / metrics / reflections /
  review-summary contracts (Sessions 6–7).

---

## What to verify when each slice is finished

Use this as the acceptance checklist when manually checking a completed
slice. Repeated Project names are allowed; Projects are distinct by ID.

### Slice 1 — Project discovery contract

- [x] As an authenticated Team Member, I receive only Projects I belong to.
- [x] As an authenticated Auditor, I receive all Projects.
- [x] A page request returns the expected page, stable ordering, and accurate
  total; page 1 and page size 15 are the defaults, and page size cannot
  exceed 30.
- [x] An empty membership list returns an empty `items` array, not `null`.
- [x] Invalid pagination returns a client error; unauthenticated callers
  cannot use the endpoint.
- [x] Backend tests and API documentation cover the response and access
  rules.

### Slice 2 — Project discovery and selection

- [x] The selector displays Projects returned by the authenticated API and
  lets me choose a different Project.
- [x] My selected Project remains selected after a page reload.
- [x] A stale or inaccessible saved Project ID is replaced with an
  accessible Project; if none are accessible, the selector shows the empty
  state and clears the saved selection.
- [x] While Projects load, the selector is disabled and loading is
  announced accessibly.
- [x] If discovery fails, I see an error and can retry; a successful retry
  restores the Project options.
- [x] Authenticated browser verification exercises live loading/selection
  and persistence, plus empty and simulated discovery-error/retry states.
- [x] Automated API, proxy, provider, and AppShell tests, frontend typecheck,
  and lint pass.

### Slice 3 — Project templates and creation

- [x] I can start Create Project from the selector, including when I have no
  accessible Projects.
- [x] Available templates load; I must explicitly select one before
  submitting.
- [x] A Team Member can create a Project with a valid name, description, and
  template; the creator has the template's creator role.
- [x] Invalid input, unavailable templates, unauthenticated access, and
  backend failures show actionable errors without pretending creation
  succeeded.
- [x] On success, the new Project becomes selected and I land on its minimal
  overview, without Sprint or Work Item UI.

### Slice 4 — Project members and roles

- [x] The shared shell shows the selected Project's real members, not fixture
  members; an empty Project shows an empty state.
- [x] Only a member holding the template's creator role can add an existing
  Team Member, change template-defined roles, or remove members.
- [x] Other Team Members and Auditors cannot perform those mutations, even
  if they call the backend directly.
- [x] An owner can assign only roles allowed by that Project's template.
- [x] The last owner cannot be removed or demoted; failed operations show an
  error and leave displayed data consistent with the server.
- [x] An owner may remove their own membership when another owner remains;
  the successful operation returns them to the Project list.
- [x] Switching Projects never shows one Project's member list as another
  Project's data.

### Slice 5 — Project-level Auditor review and feedback

- [x] An Auditor can self-assign to a Project; multiple Auditors may be
  assigned, and repeating self-assignment does not create duplicates.
- [x] An assigned Auditor can submit feedback; an unassigned Auditor cannot.
- [x] An assigned Auditor cannot unassign before the Project is permanently
  deleted.
- [x] Project Team Members can read attributed, timestamped feedback history
  but cannot submit or reply.
- [x] Feedback entries are append-only and Project-specific; they are not
  Work Item comments or Sprint reviews.
- [x] Loading, empty history, success, and API error states are usable.
- [x] Feedback fields and validation limits are agreed before this slice is
  implemented.

### Slice 6 — Project context consistency

- [x] Project overview, members, feedback, assignments, and settings always
  belong to the currently selected Project, not a seed/default Project.
- [x] When I switch Projects, the previous Project's data is cleared
  immediately while the new Project loads.
- [x] A late response from the previous Project cannot overwrite the current
  Project's view.
- [x] A stale or inaccessible selection shows a safe recovery state without
  exposing data from another Project.
- [x] After deleting the selected Project, its saved selection is cleared and
  another accessible Project is selected, or the empty state is shown.

### Slice 7 — Project settings and shell cleanup

- [x] Only a Project owner (the member with the template's creator role) can
  edit Project settings or permanently delete the Project.
- [x] The owner can edit the name and description.
- [x] Template changes succeed only when the Project has no Work Items or
  Sprints and only its creator is a member; the creator receives the new
  template's creator role.
- [x] Template changes are rejected for Projects with Work Items, Sprints,
  or additional members, and unauthorized callers are rejected by the
  backend.
- [x] Permanent deletion clearly warns that it cannot be undone and requires
  explicit confirmation; the exact confirmation interaction must be
  decided before implementation.
- [x] Successful deletion removes the Project, membership group, Auditor
  assignments, feedback, Sprints, and Work Items, while retaining user
  accounts.
- [x] If deletion cleanup fails, the UI does not report success and the
  operation can be safely retried or otherwise recovered.
- [x] There is no archive/restore flow, and the notification bell is removed.
- [x] Every remaining visible project-level control works or is clearly
  disabled.

### Slice 8 — Selected-project execution wiring

- [x] After I select Project B, Board / Backlog / Sprints request Project B’s
  id (network tab), not the seed UUID.
- [x] A brand-new Project shows empty sprints and backlog until I create
  data.
- [x] Switching Projects clears previous sprint/WI lists before new data
  appears; a late response cannot revive the old Project.
- [x] With no selection, those pages show empty/disabled — not silent seed
  data.
- [ ] Automated tests cover client project_id plumbing; typecheck/lint pass.

### Slice 9 — Overview discoverability and shell honesty

- [x] From the shell I can open the selected Project’s overview/settings
  without typing `/projects/:id`.
- [x] Board / Backlog / Sprints either follow selection (after Slice 8) or
  are clearly disabled/labeled until then.
- [x] Placeholder search / no-op buttons are removed or wired.
- [x] Reports nav and notification bell stay absent (**OUT**).

### Slice 10 — Session 1 Sprint polish

- [x] Create / edit / start / end / delete Sprint on a Project I created.
- [x] Planned points display when WIs have estimates.
- [x] Empty-start rule enforced by backend **or** clearly deferred.
- [x] Unauthorized mutations fail with a usable error.

### Slice 11 — Session 2 Board polish

- [x] Board shows the selected Project’s active Sprint only.
- [x] Status changes persist; points update.
- [x] No active Sprint → empty state → Sprints.
- [x] DnD / live sync documented as deferred.

### Slice 12 — Session 3 Backlog / WI planning

- [x] Template types + validate API used on create/edit.
- [x] AC editable for User Stories.
- [x] Blocked workflow usable; comment UI enabled in Session 5 / Slice 14.
- [x] No fake backlog reorder.

### Slice 13 — Session 4 Backend comments

- [x] OpenAPI + tests for comment create/list (+ authz).
- [x] Optional `blocked_reason` deferred — Session 3 `features.blocked_reason` remains.
- [x] Curl succeeds with member and Auditor tokens as designed.

### Slice 14 — Session 5 Frontend comments

- [x] Comments show author/time and survive reload.
- [x] Stub `listComments → []` gone.
- [x] Unauthorized users cannot post.

### Slice 15 — Session 6 Metrics / work-summary

- [x] work-summary and metrics load for a Sprint on the selected Project.
- [x] Partial data does not invent charts.

### Slice 16 — Session 7 Review / reflections / Reports

- [x] Reflection submit + list works.
- [x] Review summary preview + export works.
- [x] Reports nav points at a real page.

### Slice 17 — Session 8 Auditor observability

- [x] Auditor reviews non-seed Backlog / Board / metrics / ticket comments.
- [x] Project-level feedback (Slice 5) still works.
