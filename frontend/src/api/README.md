# API client

Typed functions that call the `backend/` Fiber API live here, one file
per resource (e.g. `projects.ts`, `sprints.ts`). Components and hooks
call into this layer instead of using `fetch` directly.
