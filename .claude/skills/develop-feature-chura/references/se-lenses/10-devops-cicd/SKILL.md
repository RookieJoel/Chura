---
name: 10-devops-cicd
description: Design and review delivery pipelines, containers, build/test automation, deployment strategies, environment configuration, infrastructure checks, rollback, observability, and release safety.
---

# DevOps and CI/CD

## Purpose

Use this skill when the task is primarily about **DevOps and CI/CD** in a software engineering project.

## Core focus

- CI pipeline
- Build / test / lint automation
- Docker
- Secrets in pipelines
- Deployment strategy
- Rollback
- Release observability

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

1. Understand environments, release cadence, and deployment target.
2. Define pipeline stages from validation to deployment.
3. Run tests and security checks before deployment.
4. Build immutable artifacts and container images when applicable.
5. Separate secrets and environment configuration from source code.
6. Define deployment, health checks, rollback, and post-deploy verification.
7. Add logging, metrics, and release observability.

## Expected outputs

- CI/CD Pipeline
- Build / Test Stages
- Docker Strategy
- Environment / Secrets Plan
- Deployment Strategy
- Rollback Plan
- Post-Deploy Verification

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
ใช้ Skill 10-devops-cicd ออกแบบหรือ Review CI/CD ของ Project นี้

ตรวจ:
- Source Control Workflow
- Build
- Lint
- Unit / Integration Test
- Security Scan
- Dependency Scan
- Artifact / Docker Build
- Secrets Management
- Environment Configuration
- Deployment
- Health Check
- Rollback
- Logging / Monitoring

ผลลัพธ์:
1. Pipeline Diagram
2. CI/CD Stages
3. Quality Gates
4. Deployment Strategy
5. Secrets / Environment Plan
6. Rollback Plan
7. Post-Deploy Verification
```
