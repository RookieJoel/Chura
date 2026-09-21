**ADR-002: Backend Framework**

| Title | Backend Framework |
| :---- | :---- |
| **Context** | Chura requires backend logic for authentication and authorization, project management, ticket assignment, backlog management, Sprint execution, progress metrics, Auditor review, and reporting. The backend should support the proposed Modular Monolith architecture with clear business-module boundaries. The development team needs a framework that is structured enough to prevent the codebase from becoming an unorganized monolith, while remaining simple to deploy during the pilot. The backend must also support server-side authorization because Auditor and Team Member permissions cannot rely only on frontend restrictions. |
| **Decision** | Use Go (Golang) as the backend language and framework foundation for Chura. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Provides static typing and compile-time checking for backend reliability.<br>• Produces a single deployable binary, simplifying deployment for the pilot.<br>• Provides strong concurrency support for handling multiple project, ticket, and dashboard requests.<br>• Has a small runtime footprint and good performance for API workloads.<br>• Supports clear package boundaries that can align with the Modular Monolith structure.<br>**Negative**<br>• The development team must be familiar with Go conventions, interfaces, error handling, and package design.<br>• Go provides less framework-level structure out of the box than highly opinionated frameworks such as NestJS, so the team must define architectural conventions carefully. |

*(Source: `proposal.md`.)*
