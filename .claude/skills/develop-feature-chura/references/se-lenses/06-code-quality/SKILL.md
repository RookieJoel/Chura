---
name: 06-code-quality
description: Review code for readability, correctness, maintainability, complexity, duplication, naming, cohesion, coupling, error handling, and adherence to project conventions without unnecessary rewrites.
---

# Code Quality

## Purpose

Use this skill when the task is primarily about **Code Quality** in a software engineering project.

## Core focus

- Readability and naming
- Complexity
- Duplication
- Coupling and cohesion
- Error handling
- Dead code
- Project consistency

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

1. Understand intent before suggesting changes.
2. Find correctness issues first, then maintainability issues.
3. Prefer small, behavior-preserving improvements.
4. Flag high complexity, duplication, hidden side effects, and unclear naming.
5. Check error handling and boundary conditions.
6. Respect existing project conventions unless they are unsafe.
7. Prioritize findings by impact and confidence.

## Expected outputs

- Code Quality Findings
- Severity / Priority
- File / Function Locations
- Refactoring Suggestions
- Before / After Examples
- Maintainability Risks

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
ใช้ Skill 06-code-quality Review Code ของ Project นี้

ให้ตรวจ:
- Correctness
- Readability
- Naming
- Function / Class Responsibility
- Cyclomatic Complexity
- Duplication
- Coupling / Cohesion
- Error Handling
- Dead Code
- Magic Values
- Side Effects
- Project Convention

แต่ละ Finding ให้มี:
1. Priority
2. File / Function
3. Problem
4. Why it matters
5. Recommended improvement
6. Example refactor

ยังไม่ต้องแก้ Code จนกว่าจะสั่ง
```
