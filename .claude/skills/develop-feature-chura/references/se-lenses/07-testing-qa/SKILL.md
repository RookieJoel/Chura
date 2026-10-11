---
name: 07-testing-qa
description: Create and review test strategies, unit tests, integration tests, API tests, end-to-end tests, acceptance tests, regression coverage, test data, and quality gates based on risk and requirements.
---

# Testing and QA

## Purpose

Use this skill when the task is primarily about **Testing and QA** in a software engineering project.

## Core focus

- Test strategy
- Unit / integration / API / E2E tests
- Acceptance criteria coverage
- Boundary and negative tests
- Regression tests
- Test data
- Quality gates

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

1. Trace tests back to requirements and risks.
2. Choose the lowest-cost test level that gives sufficient confidence.
3. Cover happy paths, boundaries, negative cases, and failure handling.
4. Avoid brittle tests that duplicate implementation details.
5. Design repeatable test data and isolation strategy.
6. Identify missing regression coverage.
7. Define measurable quality gates before release.

## Expected outputs

- Test Strategy
- Test Matrix
- Unit / Integration / API / E2E Cases
- Acceptance Coverage
- Regression Plan
- Quality Gate

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
ใช้ Skill 07-testing-qa สร้าง Test Strategy สำหรับระบบนี้

ให้ครอบคลุม:
- Unit Test
- Integration Test
- API Test
- End-to-End Test
- Acceptance Test
- Negative / Boundary Test
- Regression Test
- Test Data
- Failure Scenario

ให้ Mapping:
Requirement → Acceptance Criteria → Test Case

ผลลัพธ์:
1. Test Strategy
2. Test Matrix
3. Test Cases
4. Missing Coverage
5. Automation Candidates
6. Quality Gate ก่อน Release
```
