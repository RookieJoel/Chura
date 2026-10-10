---
name: 03-software-architecture
description: Define maintainable software architecture using separation of concerns, SOLID, clean architecture, modular design, DDD concepts, dependency rules, and documented architecture decisions.
---

# Software Architecture

## Purpose

Use this skill when the task is primarily about **Software Architecture** in a software engineering project.

## Core focus

- Architecture style selection
- Layer/module boundaries
- SOLID and separation of concerns
- Dependency direction
- Domain modeling
- Architecture decision records
- Maintainability and evolvability

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

1. Identify domain boundaries and business capabilities.
2. Choose the simplest architecture that satisfies current requirements.
3. Define layers/modules and dependency rules.
4. Separate domain logic from infrastructure concerns.
5. Define interfaces/contracts between modules.
6. Document major trade-offs with ADR-style decisions.
7. Review architecture for coupling, cohesion, testability, and changeability.

## Expected outputs

- Architecture Style
- Module / Layer Structure
- Dependency Rules
- Domain Boundaries
- Interfaces / Contracts
- ADR Decisions
- Architecture Risks

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
ใช้ Skill 03-software-architecture ออกแบบหรือ Review Architecture ของระบบนี้

ตรวจสอบ:
- Separation of Concerns
- SOLID
- Coupling / Cohesion
- Dependency Direction
- Domain / Application / Infrastructure Separation
- Module Boundaries
- Testability
- Maintainability
- Scalability ที่จำเป็นจริง

ผลลัพธ์:
1. Architecture Style ที่แนะนำ
2. Layer / Module Structure
3. Dependency Rules
4. Domain Boundaries
5. Interfaces ระหว่าง Module
6. จุดเสี่ยงของ Architecture
7. ADR สำหรับ Decision สำคัญ
8. Refactoring Plan หากเป็นระบบเดิม
```
