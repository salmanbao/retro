# Data Model Requirements Quality Checklist: Profile Enrichment and Verification

**Purpose**: Validate data model requirements quality in specification
**Created**: 2026-05-16
**Feature**: specs/003-profile-enrichment/spec.md

**Focus Areas**: Entity definitions, relationships, validation rules, state transitions
**Depth**: Standard
**Audience**: Data model reviewer

## Entity Completeness

- [X] CHK001 Are all entities defined: ProfileEnrichment, PortfolioItem, AudienceData, FollowerVerification, PayoutPreferences, KYCStatus? [Completeness, Spec §Key Entities]
- [X] CHK002 Are entity field names and types specified? [Completeness, Spec §Key Entities]
- [X] CHK003 Is social_links embedded as JSONB in ProfileEnrichment? [Traceability, Spec §Clarifications]
- [X] CHK004 Are timestamp fields (created_at, updated_at) defined for all entities? [Completeness]

## Relationship Definitions

- [X] CHK005 Are foreign key relationships to Profile explicitly defined? [Completeness, Spec §Data Model]
- [X] CHK006 Is the 1:1 relationship between Profile and each enrichment entity specified? [Clarity, Spec §Data Model]
- [X] CHK007 Is the 1:many relationship between Profile and PortfolioItem documented? [Completeness]
- [X] CHK008 Are soft-delete relationships (deleted_at) for PortfolioItem specified? [Coverage, Spec §FR-018]

## Validation Rules

- [X] CHK009 Are language code validation requirements (ISO 639-1) specified? [Completeness, Spec §FR-019]
- [X] CHK010 Are timezone validation requirements (IANA) specified? [Completeness, Spec §FR-020]
- [X] CHK011 Are country code validation requirements (ISO 3166-1 alpha-2) specified? [Assumption, Spec §Assumptions]
- [X] CHK012 Are currency code validation requirements (ISO 4217) specified? [Assumption, Spec §Assumptions]
- [X] CHK013 Is audience_demographics max size (10KB) specified? [Clarity, Spec §SC-003]
- [X] CHK014 Is display_order validation (positive integer) defined? [Completeness, Spec §Data Model]

## State Transitions

- [X] CHK015 Are FollowerVerification status values (unverified/pending/verified/rejected) defined? [Completeness, Spec §Key Entities]
- [X] CHK016 Are KYCStatus status values (not_started/pending/approved/rejected/suspended) defined? [Completeness, Spec §Key Entities]
- [X] CHK017 Is any-state-transition allowed policy documented? [Clarity, Spec §Clarifications]
- [X] CHK018 Are status_history table requirements for audit specified? [Completeness, Spec §Clarifications]

## Soft Deletion

- [X] CHK019 Is soft deletion requirement for PortfolioItem documented? [Completeness, Spec §FR-006, FR-018]
- [X] CHK020 Is the behavior that soft-deleted items return 404 on direct access specified? [Clarity, Spec §Edge Cases]
- [X] CHK021 Is the constraint that deleted items are excluded from normal queries defined? [Consistency, Spec §SC-007]

## Display Order Semantics

- [X] CHK022 Is display_order tiebreaker (created_at) for identical values specified? [Clarity, Spec §Clarifications]
- [X] CHK023 Is no auto-renumber behavior after delete specified? [Completeness, Spec §Clarifications]
- [X] CHK024 Is max portfolio items limit (50) enforced in service layer? [Traceability, Spec §NFR-003]

## Data Model Consistency

- [X] CHK025 Are encrypted_details field semantics consistent (stored but never returned)? [Consistency, Spec §FR-015]
- [X] CHK026 Are JSONB fields (social_links, demographics, evidence_urls) validation approach defined? [Gap]

## Notes

- Entity completeness checked against Key Entities section
- Relationships should map to existing Profile aggregate
- Validation rules must be testable and measurable