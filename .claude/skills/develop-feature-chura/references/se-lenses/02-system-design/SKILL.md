---
name: 02-system-design
description: Design end-to-end software systems from requirements, including boundaries, components, data flows, integrations, deployment context, scalability, availability, and architecture diagrams.
---

# System Design

## Purpose

Use this skill when the task is primarily about **System Design** in a software engineering project.

## Core focus

- System context and boundaries
- Component decomposition
- Data flow and integration
- Scalability and availability
- Caching / queues / storage
- Failure handling
- Architecture diagrams

## Operating principles

- Prefer evidence from the actual project over assumptions.
- State assumptions clearly when information is incomplete.
- Do not invent files, endpoints, requirements, test results, or security findings.
- Prefer the simplest design that satisfies the stated requirements.
- Preserve existing business behavior unless the task explicitly authorizes changes.
- When editing code, make small, reviewable changes and explain the impact.
- When possible, connect recommendations to requirements, risks, and verification steps.
- Do not claim tests or commands passed unless they were actually run.

## Workflow

1. Start from validated requirements and expected workload.
2. Define system boundary, actors, external systems, and trust boundaries.
3. Estimate scale only when useful and state assumptions.
4. Design major components and responsibilities.
5. Define synchronous/asynchronous interactions and data flow.
6. Identify bottlenecks, failure modes, and scaling strategy.
7. Produce diagrams and design trade-offs.

## Expected outputs

- System Context
- High-Level Architecture
- Component Responsibilities
- Data Flow
- Sequence Flow
- Scalability Plan
- Availability / Resilience Plan
- Trade-offs / Risks

## Quality checklist

Before finishing:
- Is the recommendation traceable to the user's goal or project evidence?
- Are assumptions and trade-offs explicit?
- Are risks prioritized?
- Are examples concrete enough to implement?
- Are verification steps included where applicable?
- Did you avoid unnecessary complexity?

## Thai quick-start prompt

```text
ใช้ Skill 02-system-design ออกแบบระบบจาก Requirement ที่มี

ให้วิเคราะห์:
- User / External System
- System Boundary
- Main Components
- Data Flow
- Integration Points
- Storage
- Cache / Queue หากจำเป็น
- Scalability
- Availability
- Failure Scenarios
- Security Boundary

ให้สร้าง:
1. System Context
2. High-Level Architecture
3. Component Diagram
4. Data Flow
5. Sequence Flow สำหรับ Use Case สำคัญ
6. Scaling Strategy
7. Failure Handling
8. Design Trade-offs

ใช้ Mermaid Diagram เมื่อเหมาะสม
```
