# Automatic Completion Behavior Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate auto-completion rules, trigger criteria, and behavioral guarantees
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md)

## Auto-Complete Trigger Criteria

^- [X] CHK001 - Is the all-or-nothing criteria for auto-completion explicitly defined? [Clarity, Spec §Clarifications]
^- [X] CHK002 - Are the four auto-complete keys enumerated (profile_enrichment, payout_preferences, kyc_status, social_links)? [Completeness, Spec §BR-007]
^- [X] CHK003 - Is the auto-completion criteria for profile_enrichment defined (bio AND avatar)? [Completeness, Spec §BR-007]
^- [X] CHK004 - Is the auto-completion criteria for payout_preferences defined (encrypted details)? [Completeness, Spec §BR-007]
^- [X] CHK005 - Is the auto-completion criteria for kyc_status defined (status = approved)? [Completeness, Spec §BR-007]
^- [X] CHK006 - Is the auto-completion criteria for social_links defined (has entries)? [Completeness, Spec §BR-007]

## Partial Data Handling

^- [X] CHK007 - Is the behavior when only bio (no avatar) is present defined? [Edge Case, Gap]
^- [X] CHK008 - Is the behavior when only avatar (no bio) is present defined? [Edge Case, Gap]
^- [X] CHK009 - Is "partial data does NOT trigger auto-completion" explicitly stated? [Clarity, Spec §BR-007]

## Completion Persistence

^- [X] CHK010 - Is the "no reversal" rule for completed steps explicitly defined? [Completeness, Spec §BR-008]
^- [X] CHK011 - Is the behavior when source data is removed after auto-completion documented? [Edge Case, Spec §BR-008]
^- [X] CHK012 - Does recalculate only add missing steps, never reverting existing completions? [Correctness, Spec §BR-008, Spec §Specmd §6]

## Recalculate Behavior

^- [X] CHK013 - Is recalculate triggered explicitly via POST endpoint? [Clarity, Spec §4]
^- [X] CHK014 - Does recalculate add missing steps from newer template versions? [Clarity, Spec §BR-006]
^- [X] CHK015 - Does recalculate preserve existing completed steps? [Clarity, Spec §BR-008]
^- [X] CHK016 - Does recalculate apply auto-completion to newly added steps? [Completeness, Gap]

## Snapshot Versioning

^- [X] CHK017 - Is the snapshot versioning approach documented? [Completeness, Spec §Clarifications]
^- [X] CHK018 - Is the behavior when template version changes documented? [Clarity, Spec §BR-006]
^- [X] CHK019 - Are existing progress records locked to their original template version? [Clarity, Spec §BR-006]

## Notes

Auto-completion behavior must be deterministic and predictable to avoid user confusion.