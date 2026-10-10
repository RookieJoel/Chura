---
name: 04-database-design
description: Design and review relational or document databases, including entities, relationships, normalization, constraints, indexes, transaction boundaries, migrations, and query-aware schema decisions.
---

# Database Design

## Purpose

Use this skill when the task is primarily about **Database Design** in a software engineering project.

## Core focus

- Entity / relationship modeling
- Normalization and denormalization
- Primary / foreign keys
- Constraints and validation
- Indexes
- Transactions and consistency
- Migration strategy

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

1. Derive entities and relationships from domain requirements.
2. Define identifiers, cardinality, ownership, and lifecycle.
3. Normalize by default; denormalize only for justified performance or access patterns.
4. Add constraints to protect data integrity.
5. Design indexes from actual query patterns.
6. Define transaction boundaries and consistency expectations.
7. Plan migrations, backups, and rollback considerations.

## Expected outputs

- ERD
- Table / Collection Definitions
- Keys / Constraints
- Indexes
- Transaction Rules
- Migration Notes
- Query / Performance Risks

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
ใช้ Skill 04-database-design ออกแบบ Database ของระบบนี้

ให้ดำเนินการ:
- ระบุ Entity และ Relationship
- สร้าง ERD
- กำหนด PK / FK
- ระบุ Cardinality
- ออกแบบ Constraints
- ตรวจ Normalization
- ออกแบบ Index ตาม Query Pattern
- ระบุ Transaction Boundary
- พิจารณา Migration / Backup / Rollback

ผลลัพธ์:
1. ERD
2. Schema
3. Data Dictionary
4. Constraints
5. Index Strategy
6. Transaction Rules
7. Migration Plan
8. Performance Risks
```
