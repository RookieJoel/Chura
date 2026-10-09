---
name: 01-requirements-engineering
description: Analyze software requirements and convert stakeholder needs into clear scope, user stories, acceptance criteria, business rules, non-functional requirements, risks, and traceable requirements.
---

# Requirements Engineering

## Purpose

Use this skill when the task is primarily about **Requirements Engineering** in a software engineering project.

## Core focus

- Stakeholder and problem analysis
- Functional / non-functional requirements
- User stories and use cases
- Acceptance criteria
- Business rules and constraints
- Requirement prioritization
- Traceability and ambiguity detection

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

1. Understand business goal, users, context, and constraints.
2. Separate problem statements from proposed solutions.
3. Identify functional requirements, non-functional requirements, business rules, and assumptions.
4. Rewrite ambiguous requirements into testable statements.
5. Create user stories/use cases and acceptance criteria.
6. Identify missing information, conflicts, dependencies, and risks.
7. Produce a traceability-ready requirements summary.

## Expected outputs

- Problem Statement
- Scope / Out of Scope
- Stakeholder Map
- Functional Requirements
- Non-Functional Requirements
- User Stories / Use Cases
- Acceptance Criteria
- Business Rules
- Assumptions / Constraints
- Requirement Traceability Table

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
ใช้ Skill 01-requirements-engineering วิเคราะห์ Requirement ของระบบนี้

เป้าหมาย:
- แยก Problem / Need / Solution ออกจากกัน
- ระบุ Functional และ Non-Functional Requirements
- เขียน User Story หรือ Use Case
- เขียน Acceptance Criteria ที่ทดสอบได้
- ระบุ Business Rules, Constraints, Assumptions และ Risks
- ตรวจหาความกำกวม ความขัดแย้ง และ Requirement ที่ตกหล่น

ผลลัพธ์สุดท้ายให้มี:
1. Problem Statement
2. Scope / Out of Scope
3. Functional Requirements
4. Non-Functional Requirements
5. User Stories / Use Cases
6. Acceptance Criteria
7. Business Rules
8. Risks / Open Questions
9. Requirement Traceability Table
```
