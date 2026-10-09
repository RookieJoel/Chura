---
name: 09-owasp-security
description: Review software security using OWASP-oriented secure engineering practices, covering access control, authentication, input validation, cryptography, secrets, dependencies, configuration, logging, APIs, and secure release checks.
---

# OWASP Security

## Purpose

Use this skill when the task is primarily about **OWASP Security** in a software engineering project.

## Core focus

- Access control
- Authentication and session security
- Input validation and injection
- Secrets and cryptography
- Dependencies and supply chain
- Security configuration
- Logging and incident visibility

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

1. Identify assets, trust boundaries, entry points, and sensitive data.
2. Review authentication and authorization before lower-impact issues.
3. Inspect input handling, queries, uploads, and serialization.
4. Check secrets, cryptography, dependency risks, and configuration.
5. Review logging, error handling, and monitoring for abuse detection.
6. Map findings to OWASP categories when appropriate.
7. Rank by severity, provide evidence, remediation, and verification steps.

## Expected outputs

- Security Findings
- Severity
- Evidence / File Locations
- OWASP Mapping
- Remediation
- Verification
- Residual Risks

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
ใช้ Skill 09-owasp-security ตรวจ Security ของ Project นี้

ตรวจ:
- Authentication
- Authorization / RBAC
- API Security
- Input Validation
- Injection
- Database Access
- File Upload
- Session / Token
- Secrets
- Cryptography
- Dependencies
- Configuration
- Logging
- Error Handling
- CI/CD

แต่ละ Finding ต้องมี:
1. Severity
2. Evidence
3. File / Function
4. Impact
5. OWASP Mapping
6. Recommended Fix
7. Verification

ขั้นแรก Review อย่างเดียว ยังไม่แก้ Code
```
