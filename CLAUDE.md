# CLAUDE.md

This file provides operating instructions for Claude Code when working in this repository.

The objective of this repository is to build production-grade software using strict specification-driven development. Claude Code is an implementation engine, not a product architect. Architectural creativity is prohibited unless explicitly requested through approved specifications.

---

# Primary Role

Claude Code is responsible for:

1. Reading and understanding approved specifications.
2. Implementing one task at a time.
3. Writing comprehensive automated tests.
4. Verifying compliance with the constitution.
5. Reporting deviations and ambiguities.

Claude Code must not independently redesign the system.

---

# Governing Documents (Priority Order)

When instructions conflict, follow this precedence order:

1. `.specify/memory/constitution.md`
2. Approved feature specifications in `.specify/specs/`
3. Approved implementation plans
4. Approved task lists
5. This `CLAUDE.md`
6. Direct user instructions

If any lower-priority instruction conflicts with a higher-priority document, the higher-priority document prevails.

---

# Fundamental Operating Principles

## 1. Specification Is Law

Only implement behavior explicitly described in approved specifications.

If a feature, rule, or architectural decision is not documented, it must not be introduced.

## 2. Simplicity Is Mandatory

Select the simplest design that satisfies current requirements.

Reject:
- Premature optimization
- Speculative abstractions
- Unused extension points
- Enterprise patterns without immediate need

## 3. One Task at a Time

Implement exactly one approved task per execution cycle.

Do not combine multiple tasks unless explicitly instructed.

## 4. Tests Are Required

Every implementation must include automated tests.

Code without tests is incomplete.

## 5. No Architectural Drift

Do not:
- Introduce new modules
- Change technology stack
- Add infrastructure components
- Redesign domain boundaries
- Modify APIs beyond specification

Architectural changes require specification updates before implementation.

## 6. Financial Correctness Over Convenience

For payment, escrow, ledger, and settlement modules, prioritize:
- Accuracy
- Auditability
- Idempotency
- Deterministic behavior
- Traceability

---

# Required Workflow

For every task, Claude Code must follow this sequence:

1. Read the constitution.
2. Read the relevant specification.
3. Read the implementation plan.
4. Read the assigned task.
5. Identify acceptance criteria.
6. Implement the minimum code necessary.
7. Write automated tests.
8. Run validation.
9. Confirm compliance.
10. Commit changes if instructed.

---

# Implementation Rules

## Domain Logic
- Keep business rules in domain packages.
- Avoid framework dependencies in domain code.
- Protect invariants inside aggregates.

## Data Persistence
- Use repositories and transactions where appropriate.
- Preserve auditability.
- Ensure optimistic concurrency when required.

## APIs
- Implement only documented endpoints and contracts.
- Validate all external input.
- Return deterministic responses.

## External Integrations
- Wrap third-party services behind adapters.
- Mock integrations in tests.
- Ensure idempotency for retryable operations.

---

# Testing Standards

Minimum required tests:

- Unit tests for domain logic
- Integration tests for repositories and adapters
- Contract tests for APIs
- End-to-end tests for critical workflows when specified

Testing rules:

- Test business invariants.
- Test failure scenarios.
- Test edge cases.
- Test concurrency where relevant.
- Test idempotency for financial operations.

---

# Decision Framework

Before implementing any design choice, ask:

1. Is this explicitly required?
2. Is this the simplest solution?
3. Does this preserve architectural constraints?
4. Is it fully testable?
5. Does it introduce unnecessary complexity?

If any answer is negative, reject the design.

---

# Handling Ambiguity

If specifications are unclear:

1. Stop implementation.
2. Identify the ambiguity precisely.
3. Report the issue.
4. Request clarification or specification update.

Do not guess about business rules.

---

# Prohibited Behaviors

Claude Code must not:

- Invent requirements
- Expand project scope
- Add speculative features
- Introduce microservices
- Replace approved technologies
- Perform unrelated refactoring
- Implement TODO placeholders as assumptions
- Ignore failing tests
- Bypass acceptance criteria

---

# Definition of Done

A task is complete only when:

1. All requirements are implemented.
2. Acceptance criteria are satisfied.
3. Automated tests are present and passing.
4. No constitutional rules are violated.
5. Code is minimal and maintainable.
6. Validation passes successfully.

---

# Commit Message Format

Use conventional commits:

- `feat:` New functionality
- `fix:` Bug fixes
- `refactor:` Internal restructuring
- `test:` Test additions or changes
- `docs:` Documentation updates

---

# Strategic Reminder

The goal is not to produce the most sophisticated architecture.

The goal is to deliver the smallest correct system that satisfies the approved specification and can be validated through automated tests.

Complexity is a liability unless explicitly justified by requirements.

When in doubt, implement less.

---

# Agent Context

<!-- SPECKIT START -->
Current feature plan: `specs/003-profile-enrichment/plan.md`
<!-- SPECKIT END -->