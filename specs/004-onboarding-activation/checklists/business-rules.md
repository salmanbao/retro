# Business Rules Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate business rule clarity, consistency, and completeness
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md)

## Rule Completeness

^- [X] CHK001 - Are all state transitions explicitly defined? [Completeness, Spec §FR-009]
^- [X] CHK002 - Is the criteria for each activation state transition documented? [Clarity, Spec §FR-009]
^- [X] CHK003 - Are the allowed step status transitions enumerated? [Completeness, Spec §2.2]
^- [X] CHK004 - Is the "required step cannot be skipped" rule documented? [Clarity, Spec §BR-003]
^- [X] CHK005 - Is the "optional step may be skipped" rule documented? [Clarity, Spec §BR-004]
^- [X] CHK006 - Are the auto-completion trigger criteria (all-or-nothing) explicitly defined? [Clarity, Spec §BR-007]
^- [X] CHK007 - Are the conditions for each profile type's activation fully specified? [Completeness, Spec §FR-012-014]

## Rule Consistency

^- [X] CHK008 - Is the activation state machine consistent across all profile types? [Consistency, Spec §FR-009]
^- [X] CHK009 - Do step types map consistently to auto-complete keys? [Consistency, Spec §BR-007]
^- [X] CHK010 - Are business rules aligned between spec and data model validation rules? [Consistency, Spec §V001-V005]
^- [X] CHK011 - Does the "no reversal" rule apply uniformly across all step types? [Consistency, Spec §BR-008]

## Authorization Rules

^- [X] CHK012 - Is the ownership verification requirement documented? [Completeness, Spec §BR-002]
^- [X] CHK013 - Is the admin-only activation approval requirement documented? [Completeness, Spec §BR-005]
^- [X] CHK014 - Are the consequences of marking a required step as skipped defined? [Clarity, Spec §BR-003]

## Auto-Completion Rules

^- [X] CHK015 - Is the all-or-nothing criteria for profile_completion explicitly stated? [Clarity, Spec §BR-007]
^- [X] CHK016 - Is the all-or-nothing criteria for payout_preferences explicitly stated? [Clarity, Spec §BR-007]
^- [X] CHK017 - Is the all-or-nothing criteria for kyc_status explicitly stated? [Clarity, Spec §BR-007]
^- [X] CHK018 - Is the all-or-nothing criteria for social_links explicitly stated? [Clarity, Spec §BR-007]

## Notes

All business rules should be testable and traceable to acceptance criteria.