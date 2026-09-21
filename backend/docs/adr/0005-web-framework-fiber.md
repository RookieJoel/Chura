**ADR-005: HTTP Framework**

| Title | HTTP Framework |
| :---- | :---- |
| **Context** | ADR-002 fixes Go as the backend language but leaves the HTTP framework open. Candidates considered: Fiber, Gin, Echo, stdlib `net/http`. The backend sits behind `internal/adapter/http` in the hexagonal layout (ADR-004), so the framework choice is isolated to that one package. |
| **Decision** | Use Fiber v2 for the `internal/adapter/http` adapter. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Fast, Express-like ergonomics, low boilerplate for route grouping/middleware (useful for Team Member vs Auditor role gating).<br>• Good fit for a small student team.<br>**Negative**<br>• Built on `fasthttp`, not stdlib `net/http` — some third-party middleware assumes stdlib and needs adapting.<br>• Another framework choice the team must learn. |
