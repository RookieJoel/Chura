# Remaining Project Work (excluding Sprints and Work Items)

This backlog translates the project setup, membership, and Auditor project-level
feedback needs in `proposal.md` and `service.md` into dependency-ordered,
demoable slices. It is not a claim that this captures every possible project
requirement. Items that need a product decision are explicitly marked instead
of being treated as agreed features.

## In scope

- Project discovery, selection, templates, and onboarding.
- Project membership and roles.
- Project-level Auditor overview and feedback.
- Project settings and permanent deletion.
- Shared project navigation needed for these capabilities.

## Explicitly out of scope

Sprint creation, planning, execution, analytics, and reports; Work Item
creation, editing, assignment, comments, search, and filtering; project views
whose implementation would require those features. Permanent Project
deletion must remove associated Sprint and Work Item records, but does not add
or change Sprint or Work Item functionality.

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
AppShell integration tests pass (4 files, 10 tests). Frontend lint and
typecheck pass. Authenticated browser verification confirmed the live
Project list loads and selection persists after reload; empty and 503
discovery responses showed the expected states, and retry recovered. Vitest
requires Node 22.19 or newer in this environment; checks were run with Node
22.20. The production build compiles but cannot prerender `/board` without
the Work Item WebSocket service; that dependency is outside this slice.

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

**Done:** Template list/detail and Project creation use authenticated
same-origin proxy routes. A Team Member can create from an explicitly chosen
template, sees backend errors, and lands on the new Project's minimal
overview with the selection persisted. The overview omits Sprint and Work
Item controls. The 10-file frontend test suite (28 tests), typecheck, lint,
and production build pass; lint reports one pre-existing unused-import
warning elsewhere. Browser verification exercised the create-to-overview
flow with intercepted create/detail responses to avoid writing a test
Project to the local database.

### 4. Project members and roles — vertical slice

**Blocked by:** Slices 2 and 3 for selected-project context, the Project
overview, and template-defined role options  
**Source:** Project Configuration use case and User Access & Role Management
requirements.

- [ ] Confirm the existing authenticated `GET /api/v1/projects/:id` response
  provides the member fields needed by the UI; it already includes project
  members.
- [ ] Add same-origin authenticated proxy routes and typed clients for
  project detail, adding a member, assigning/updating a role, and removing a
  member; test authentication, request forwarding, and backend errors.
- [ ] Use the selected Project's template definition to offer only its
  allowed roles.
- [ ] Replace fixture-backed member data in the TopNav and sidebar with the
  selected project's members.
- [ ] On the Project overview, allow only a current Project owner (a member
  holding the template's creator role) to add an existing Team Member by
  email, assign/update template-defined roles, or remove a member.
- [ ] Add or update backend authorization and routes so membership changes
  enforce owner-only access; prevent removing or demoting the last owner.
- [ ] Keep ordinary members at the template's default member role unless an
  owner assigns another role allowed by that template.
- [ ] Test allowed and denied membership/role changes, including Auditor
  attempts, non-member attempts, invalid roles, and attempts to remove or
  demote the last owner.
- [ ] Cover loading, empty, and error states without showing fixtures as live
  project data.

**Done when:** Members shown in the shared shell match the selected project,
and only a Project owner can manage membership and roles; unauthorized
changes are rejected by the backend.

### 5. Project-level Auditor review and feedback — vertical slice

**Blocked by:** Slices 1 and 3 for Auditor-visible discovery and the
Project overview. Slice 7 must later remove this slice's assignments and
feedback during permanent deletion.  
**Source:** Auditor Review & Feedback requirements in `proposal.md`.

- [ ] Add Project Auditor self-assignment, allowing multiple Auditors per
  Project; an assigned Auditor cannot unassign until the Project is
  permanently deleted.
- [ ] Attach feedback to the Project itself; exclude comments on Work Items
  and Sprint reviews.
- [ ] Decide the feedback entry fields and validation limits before defining
  the persistence and API contract; do not assume a rating or fixed form.
- [ ] Allow only assigned Auditors to submit feedback; allow Team Members
  belonging to the Project to read it, but not reply.
- [ ] Add backend persistence that records feedback author and timestamp,
  preserves append-only feedback history, and stores Project Auditor
  assignments.
- [ ] Add authenticated create/read API operations, response/error contracts,
  and authorization tests for assignment, reassignment prohibition,
  Auditor submission, and member read-only access. Repeated self-assignment
  must not create duplicate assignments.
- [ ] Build an Auditor project overview limited to project-level information
  that does not depend on Sprint or Work Item functionality.
- [ ] Build UI for Auditors to self-assign and submit feedback and for
  Project members to read feedback history; cover loading, empty, success,
  and error states.

**Done when:** Assigned Auditors can leave attributed, timestamped,
append-only Project feedback; Project members can read the history; and
Auditor assignment and feedback access rules are enforced by the backend.

### 6. Project context consistency

**Blocked by:** Slices 2–5 and 7, so all project-scoped views and mutations
can share and verify the selected Project context  
**Source:** Project selection and membership use cases.

- [ ] Ensure project-scoped data (overview, members, feedback, Auditor
  assignments, and settings) uses the selected Project rather than a
  development seed ID.
- [ ] On Project switch, immediately clear the previous Project's data, show
  loading, then show the selected Project's data or a retryable error.
- [ ] Prevent late responses for a previous selection from repopulating
  stale Project data; handle stale or inaccessible selections safely.
- [ ] After deleting the selected Project, clear its persisted selection and
  select another accessible Project or show the empty state.

**Done when:** Changing the selected Project changes all in-scope project
metadata and member data without leaking information from the previous
selection.

### 7. Scope decisions and shell cleanup

**Blocked by:** Slices 3–5 for Project creation, owner/member rules, and
feedback/assignment cleanup;
independent of the implementation slices for notification cleanup

- [ ] Add authenticated backend update/delete operations and same-origin
  authenticated frontend proxies/typed clients for Project settings and
  deletion.
- [ ] Add Project name and description editing, restricted to a Project owner
  (the member holding the template's creator role).
- [ ] Allow template changes only when the Project has no Work Items or
  Sprints and only its creator is a member. Ensure that owner receives the
  new template's creator role; no other member roles need mapping under this
  eligibility rule. Enforce eligibility and ownership in the backend, not
  only in the UI.
- [ ] Add permanent Project deletion for the Project owner. Delete all
  Project-owned data, including associated membership groups, Auditor
  assignments, feedback, Sprints, and Work Items; retain user accounts.
- [ ] Define and test failure handling across database deletion and identity
  provider group deletion so partial cleanup is surfaced and can be safely
  retried; never report deletion success while Project-owned data remains.
- [ ] Test unauthorized deletion/edits and template-change rejection for
  Projects with Work Items, Sprints, or additional members.
- [ ] Require an explicit confirmation before permanent deletion and ensure
  the UI explains that deletion cannot be undone. Confirm the exact
  confirmation interaction before implementation.
- [ ] Do not implement archive or restore. Permanently deleted Projects and
  their associated data are not available in the application.
- [ ] Remove the notification bell until notification behavior and an API
  scope are defined.
- [ ] Review remaining project-level navigation and controls so each
  navigates, performs an action, or is clearly disabled.

**Done when:** Project editing and permanent deletion follow the owner and
data-removal rules above, and no visible project-level control falsely
implies unsupported functionality.

## Cross-slice quality checks

- [ ] Map each agreed requirement to at least one acceptance criterion and
  automated test.
- [ ] Cover role boundaries, cross-project access, invalid input, empty data,
  and backend/proxy failures at the relevant API/UI seams.
- [ ] Verify permanent deletion removes all Project-owned records and
  membership groups while retaining user accounts; verify failed cleanup
  does not appear successful.
- [ ] Verify each completed slice end to end; do not treat a hidden UI control
  as authorization.
- [ ] Update API documentation when backend contracts are added or changed.

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

- [ ] The shared shell shows the selected Project's real members, not fixture
  members; an empty Project shows an empty state.
- [ ] Only a member holding the template's creator role can add an existing
  Team Member, change template-defined roles, or remove members.
- [ ] Other Team Members and Auditors cannot perform those mutations, even
  if they call the backend directly.
- [ ] An owner can assign only roles allowed by that Project's template.
- [ ] The last owner cannot be removed or demoted; failed operations show an
  error and leave displayed data consistent with the server.
- [ ] Switching Projects never shows one Project's member list as another
  Project's data.

### Slice 5 — Project-level Auditor review and feedback

- [ ] An Auditor can self-assign to a Project; multiple Auditors may be
  assigned, and repeating self-assignment does not create duplicates.
- [ ] An assigned Auditor can submit feedback; an unassigned Auditor cannot.
- [ ] An assigned Auditor cannot unassign before the Project is permanently
  deleted.
- [ ] Project Team Members can read attributed, timestamped feedback history
  but cannot submit or reply.
- [ ] Feedback entries are append-only and Project-specific; they are not
  Work Item comments or Sprint reviews.
- [ ] Loading, empty history, success, and API error states are usable.
- [ ] Feedback fields and validation limits are agreed before this slice is
  implemented.

### Slice 6 — Project context consistency

- [ ] Project overview, members, feedback, assignments, and settings always
  belong to the currently selected Project, not a seed/default Project.
- [ ] When I switch Projects, the previous Project's data is cleared
  immediately while the new Project loads.
- [ ] A late response from the previous Project cannot overwrite the current
  Project's view.
- [ ] A stale or inaccessible selection shows a safe recovery state without
  exposing data from another Project.
- [ ] After deleting the selected Project, its saved selection is cleared and
  another accessible Project is selected, or the empty state is shown.

### Slice 7 — Project settings and shell cleanup

- [ ] Only a Project owner (the member with the template's creator role) can
  edit Project settings or permanently delete the Project.
- [ ] The owner can edit the name and description.
- [ ] Template changes succeed only when the Project has no Work Items or
  Sprints and only its creator is a member; the creator receives the new
  template's creator role.
- [ ] Template changes are rejected for Projects with Work Items, Sprints,
  or additional members, and unauthorized callers are rejected by the
  backend.
- [ ] Permanent deletion clearly warns that it cannot be undone and requires
  explicit confirmation; the exact confirmation interaction must be
  decided before implementation.
- [ ] Successful deletion removes the Project, membership group, Auditor
  assignments, feedback, Sprints, and Work Items, while retaining user
  accounts.
- [ ] If deletion cleanup fails, the UI does not report success and the
  operation can be safely retried or otherwise recovered.
- [ ] There is no archive/restore flow, and the notification bell is removed.
- [ ] Every remaining visible project-level control works or is clearly
  disabled.
