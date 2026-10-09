---
name: 08-debugging-refactoring
description: Diagnose software defects systematically, isolate root causes, propose minimal fixes, add regression tests, and refactor code safely while preserving behavior.
---

# Debugging and Refactoring

## Purpose

Use this skill when the task is primarily about **Debugging and Refactoring** in a software engineering project.

## Core focus

- Reproduction
- Evidence gathering
- Hypothesis testing
- Root cause analysis
- Minimal fix
- Regression tests
- Behavior-preserving refactoring

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

1. Reproduce the issue reliably before editing code.
2. Collect logs, inputs, stack traces, and environmental differences.
3. Form ranked hypotheses and test them one by one.
4. Locate root cause rather than patching symptoms.
5. Implement the smallest safe fix.
6. Add regression tests that fail before and pass after the fix.
7. Refactor separately from the functional fix when possible.

## Expected outputs

- Reproduction Steps
- Evidence
- Root Cause
- Fix Plan
- Regression Test
- Refactoring Opportunities
- Verification Result

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
ใช้ Skill 08-debugging-refactoring วิเคราะห์ Bug นี้

ห้ามเดาสุ่มแก้ Code ทันที

ให้ทำตามลำดับ:
1. Reproduce
2. Collect Evidence
3. Form Hypotheses
4. Isolate
5. Identify Root Cause
6. Propose Minimal Fix
7. Add Regression Test
8. Verify
9. Refactor เพิ่มเติมเฉพาะเมื่อปลอดภัย

รายงาน:
- Root Cause
- Evidence
- Affected Files
- Fix
- Regression Test
- Risk of Change
```
