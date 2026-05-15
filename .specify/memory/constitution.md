# ECM Marketplace Constitution

## Core Principles

### I. MVP-First Development
The primary objective is to deliver the smallest production-capable system that validates business value.
Only requirements explicitly approved for the current milestone may be implemented.
All speculative features are excluded.

### II. Simplicity Over Sophistication
The simplest design that satisfies current requirements must be selected.
Premature optimization, unnecessary abstractions, and enterprise-grade patterns are prohibited unless justified by an approved requirement.

### III. Specification as the Single Source of Truth
All implementation decisions must originate from approved specifications.
Conversational suggestions, assumptions, and ad hoc design ideas have no authority unless incorporated into the specification.

### IV. Test-First Development (Non-Negotiable)
Every feature must be implemented with automated tests.
Acceptance criteria determine completion.
Code without adequate tests is incomplete.

### V. Controlled Architectural Evolution
Architecture may only change through explicit specification updates.
Implementation must not introduce new modules, technologies, or patterns that are not defined in the approved architecture.

## Architectural Constraints

### System Architecture
- Modular monolith architecture.
- Single deployable backend application.
- Clear domain boundaries.
- Domain-driven design principles.
- Clean architecture with domain isolation.

### Technology Stack
- Backend: Go
- Database: PostgreSQL
- Frontend: Next.js
- API: REST/JSON
- Containerization: Docker
- CI/CD: GitHub Actions

### Financial Integrity Requirements
- Immutable audit trail for all financial state transitions.
- Double-entry ledger for money movement.
- Idempotent payment and payout operations.
- Optimistic concurrency control where required.
- Full traceability of balances and settlements.

## Engineering Standards

### Code Quality
- Domain logic must not depend on frameworks.
- Public interfaces require tests.
- Structured logging is mandatory.
- Configuration must be environment-driven.
- Backward compatibility must be preserved unless explicitly approved.

### Testing Requirements
- Unit tests for domain logic.
- Integration tests for repositories and external adapters.
- Contract tests for APIs.
- End-to-end tests for critical business workflows.

### Security Requirements
- Principle of least privilege.
- Secrets must never be hardcoded.
- Sensitive operations require explicit authorization checks.
- Input validation is mandatory at all external boundaries.

## Development Workflow

### Specification Workflow
1. Product requirements are defined.
2. Technical plan is created.
3. Acceptance criteria are finalized.
4. Tasks are generated.
5. Implementation begins.
6. Validation confirms compliance.

### Task Execution Rules
- Only one task may be implemented at a time.
- Each task must reference specific acceptance criteria.
- Completion requires all tests to pass.
- Deviations require specification updates before coding continues.

## Decision Rules

1. If a feature is not in the specification, it does not exist.
2. If implementation conflicts with the specification, the specification prevails.
3. If multiple designs are viable, the simplest is mandatory.
4. If complexity cannot be justified by current requirements, it must be rejected.
5. If a new idea emerges during implementation, defer it unless it blocks approved requirements.

## Governance

This constitution is the highest authority in the repository.
All specifications, plans, tasks, and code must comply with it.
Any amendment requires explicit documentation and version updates.
Non-compliant code must be rejected and refactored.

**Version**: 1.0.0
**Ratified**: 2026-05-14
**Last Amended**: 2026-05-14