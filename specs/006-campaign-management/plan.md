# Implementation Plan: Campaign Management

**Branch**: `main` | **Date**: 2026-05-18 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-campaign-management/spec.md`

## Summary

Build the Campaign Management module for ViralForge allowing Brand profiles to create, manage, and publish short-form video content campaigns. The module covers campaign lifecycle from draft through cancellation, with publishing readiness validation ensuring KYC approval, onboarding completion, and payout configuration before campaigns go live.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: chi/v5 (router), GORM (PostgreSQL ORM), existing ViralForge backend stack

**Storage**: PostgreSQL (existing)

**Testing**: go test, integration tests against PostgreSQL, contract tests for HTTP endpoints

**Target Platform**: Linux server (existing backend)

**Project Type**: backend-service/web-api

**Performance Goals**: Campaign lifecycle transitions < 5s; campaign creation < 60s end-to-end

**Constraints**: Must integrate with existing auth, profile, KYC, onboarding, and payout modules

**Scale/Scope**: Single-brand multi-campaign scenarios; up to hundreds of campaigns per brand

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Test-First Development**: All functional requirements MUST have failing tests written before implementation
- [x] **Specification Law**: Only implement behavior explicitly described in approved specifications
- [x] **Simplicity**: Select simplest design that satisfies current requirements
- [x] **No Architectural Drift**: No new modules, no technology stack changes, no API modifications beyond specification

**Status**: PASS — No constitutional violations detected.

## Project Structure

### Documentation (this feature)

```text
specs/006-campaign-management/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
backend/
├── src/
│   ├── domain/           # Campaign domain models and errors
│   ├── repository/       # Campaign repository interface + GORM implementation
│   ├── service/          # Campaign service (business logic, lifecycle, validation)
│   ├── handler/          # Campaign HTTP handlers
│   │   └── campaign/
│   │       └── campaign_handler.go
│   └── middleware/       # Existing auth/ownership middleware (reuse)
└── tests/
    └── integration/      # Campaign integration + contract tests
```

**Structure Decision**: Campaign module follows existing ViralForge backend patterns. New directories: `domain/campaign.go`, `repository/campaign_repo.go`, `service/campaign_svc.go`, `handler/campaign_handler.go`. Existing middleware (auth, ownership, profile-type) is reused.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | - | - |

## Phase 0: Research

**Research artifacts**: None required — all technical decisions derivable from existing codebase patterns and Go/PostgreSQL best practices. The module uses established patterns: GORM models, chi handlers, service layer with domain errors.

## Phase 1: Design & Contracts

### Data Model

**Campaign Entity** (in `domain/campaign.go`):
- Core: `id` (UUID), `brand_profile_id` (FK), `title`, `slug`, `summary`, `description`, `objective`, `product_name`, `landing_url`
- Budget: `total_budget` (decimal), `currency` (string ISO 4217), `target_clips` (int), `target_posts` (int), `cpv` (decimal), `min_payout` (decimal, nullable), `max_payout` (decimal, nullable)
- Timeline: `submission_start` (datetime), `submission_deadline` (datetime), `distribution_start` (datetime), `campaign_end` (datetime)
- Eligibility: `allowed_countries` ([]string), `allowed_languages` ([]string), `min_followers` (int), `platforms` ([]string), `creator_categories` ([]string)
- Creative: `min_duration_secs` (int), `max_duration_secs` (int), `aspect_ratio` (string), `talking_points` ([]string), `prohibited_claims` ([]string), `hashtags` ([]string), `cta_instructions` (string)
- Lifecycle: `status` (enum: draft/published/active/paused/completed/cancelled), `version` (int, optimistic locking)
- Audit: `created_at`, `updated_at`, `deleted_at` (soft delete)

**CampaignAsset Entity** (in `domain/campaign_asset.go`):
- `id` (UUID), `campaign_id` (FK), `url`, `asset_type` (enum: reference/raw_media/document), `description` (nullable)

**State Transitions**:
```
draft → published (via publish action, requires readiness validation)
published → active (automatic on submission_deadline reached)
active → paused (via pause action)
paused → active (via resume action)
active → completed (via complete action, or automatic on campaign_end)
draft|published|active|paused → cancelled (via cancel action, soft delete)
```

### API Contracts

**Endpoints** (all under `/api/v1/campaigns`):

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/` | Create campaign | Brand profile |
| GET | `/` | List campaigns (paginated, filterable by status) | Brand profile |
| GET | `/{id}` | Get campaign details | Brand profile (owner) |
| PATCH | `/{id}` | Update campaign (restricted by status) | Brand profile (owner) |
| POST | `/{id}/publish` | Publish campaign | Brand profile (owner) |
| POST | `/{id}/pause` | Pause campaign | Brand profile (owner) |
| POST | `/{id}/resume` | Resume campaign | Brand profile (owner) |
| POST | `/{id}/complete` | Complete campaign | Brand profile (owner) |
| POST | `/{id}/cancel` | Cancel campaign (soft delete) | Brand profile (owner) |

**Request/Response shapes** documented in `contracts/campaign-api.md`

### Agent Context Update

`CLAUDE.md` plan reference updated to point to `specs/006-campaign-management/plan.md`