# Activation Logic Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate activation state machine correctness and completeness
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md)

## State Machine Definition

^- [X] CHK001 - Are all four activation states enumerated? [Completeness, Spec §FR-008]
^- [X] CHK002 - Is the initial state (not_started) documented? [Completeness, Spec §FR-008]
^- [X] CHK003 - Is the transition trigger for not_started → onboarding documented? [Clarity, Spec §FR-009]
^- [X] CHK004 - Is the transition trigger for onboarding → pending_review documented? [Clarity, Spec §FR-009]
^- [X] CHK005 - Is the transition trigger for pending_review → activated documented? [Clarity, Spec §FR-009]
^- [X] CHK006 - Are there any invalid or disallowed state transitions defined? [Completeness, Gap]

## State Transition Correctness

^- [X] CHK007 - Does any step status change trigger not_started → onboarding? [Correctness, Spec §FR-009]
^- [X] CHK008 - Does completion of all required steps (or optional skipped) trigger onboarding → pending_review? [Correctness, Spec §FR-009]
^- [X] CHK009 - Is explicit admin approval required for pending_review → activated? [Correctness, Spec §FR-009]
^- [X] CHK010 - Can activated state be reached without admin approval? [Ambiguity, Gap]

## Role-Specific Activation

^- [X] CHK011 - Are Brand activation requirements fully specified? [Completeness, Spec §FR-012]
^- [X] CHK012 - Are Editor activation requirements fully specified? [Completeness, Spec §FR-013]
^- [X] CHK013 - Are Influencer activation requirements fully specified? [Completeness, Spec §FR-014]
^- [X] CHK014 - Do all role-specific flows converge to the same terminal state (activated)? [Consistency, Spec §FR-008]

## Edge Cases

^- [X] CHK015 - Is the behavior when no steps are started defined? [Edge Case, Spec §FR-008]
^- [X] CHK016 - Is the behavior when only optional steps are completed defined? [Edge Case, Gap]
^- [X] CHK017 - Is the behavior when all required steps are skipped (but required=true) defined? [Edge Case, Gap]
^- [X] CHK018 - Can activation be reverted once achieved? [Edge Case, Gap]

## Notes

Activation logic is the core business rule driving user progression through the platform.