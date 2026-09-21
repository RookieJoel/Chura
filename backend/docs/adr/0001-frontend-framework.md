**ADR-001: Frontend Framework**

| Title | Frontend Framework |
| :---- | :---- |
| **Context** | Chura requires a responsive web interface for Team Members and Auditors. The frontend must support project creation, ticket and backlog management, Sprint boards, progress visualization, Auditor review, and Sprint Review Summary workflows. The project is in an early pilot stage, so the team needs a framework that supports fast development, reusable UI components, maintainable routing, and future extension without introducing unnecessary infrastructure complexity. |
| **Decision** | Use Next.js with React as the frontend framework for Chura. |
| **Status** | **Accepted** |
| **Consequences** | **Positive**<br>• Provides a structured React framework with built-in routing and application conventions.<br>• Supports responsive web application development for both Team Member and Auditor workflows.<br>• Allows reuse of the React ecosystem for boards, forms, charts, and other project-management UI components.<br>• Supports future server-side rendering or server-side data fetching if the product later needs them.<br>• Keeps the frontend technology widely understandable for student developers.<br>**Negative**<br>• Introduces Next.js-specific concepts that the development team must learn.<br>• The team must define clear boundaries between frontend state, API state, and backend business rules.<br>• Some Next.js capabilities may be unnecessary during the first pilot if the application behaves mainly as an authenticated dashboard. |

*(Source: `proposal.md`.)*
