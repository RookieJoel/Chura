---
name: 05-api-design
description: Design consistent, secure, evolvable APIs with clear resources, contracts, validation, errors, pagination, idempotency, versioning, authentication, and OpenAPI-friendly definitions.
---

# API Design

## Purpose

Use this skill when the task is primarily about **API Design** in a software engineering project.

## Core focus

- RESTful resource design
- Request / response contracts
- Validation
- Error model
- Pagination / filtering / sorting
- Idempotency
- Versioning and compatibility

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

1. Model API around domain resources and use cases.
2. Define endpoints, methods, request/response schemas, and status codes.
3. Validate input explicitly and consistently.
4. Create a standard error response model.
5. Add pagination/filtering/sorting only where needed.
6. Define authentication, authorization, rate limits, and idempotency where relevant.
7. Review backwards compatibility and document with OpenAPI-compatible structure.

## Expected outputs

- Endpoint Catalog
- Request / Response Schemas
- Status Codes
- Error Contract
- Authentication / Authorization
- Pagination / Filtering
- OpenAPI Outline

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
ใช้ Skill 05-api-design ออกแบบหรือ Review API ของระบบนี้

ตรวจสอบ:
- Resource Naming
- HTTP Method
- URL Structure
- Request / Response
- Validation
- Status Code
- Error Format
- Pagination / Filter / Sort
- Authentication
- Authorization
- Rate Limit
- Idempotency
- API Versioning

ให้ส่งออก:
1. Endpoint Catalog
2. Request / Response Example
3. Error Contract
4. Auth Requirement
5. Validation Rules
6. OpenAPI Outline
7. Compatibility / Security Risks
```
