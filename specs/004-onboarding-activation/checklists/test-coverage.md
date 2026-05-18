# Test Coverage Expectations Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate that test coverage areas are defined for critical paths
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md), [quickstart.md](../quickstart.md)

## State Transition Tests

^- [X] CHK001 - Is there a test for not_started → onboarding transition? [Coverage, quickstart.md Scenario 1]
^- [X] CHK002 - Is there a test for onboarding → pending_review transition? [Coverage, quickstart.md Scenario 1]
^- [X] CHK003 - Is there a test for pending_review → activated transition? [Coverage, quickstart.md Scenario 8]
^- [X] CHK004 - Is there a test for all possible invalid state transitions? [Coverage, Gap]

## Step Status Tests

^- [X] CHK005 - Is there a test for not_started → in_progress transition? [Coverage, quickstart.md Scenario 1]
^- [X] CHK006 - Is there a test for in_progress → completed transition? [Coverage, quickstart.md Scenario 1]
^- [X] CHK007 - Is there a test for required step skip blocked (403)? [Coverage, quickstart.md Scenario 3]
^- [X] CHK008 - Is there a test for optional step skip allowed? [Coverage, quickstart.md Scenario 4]

## Auto-Completion Tests

^- [X] CHK009 - Is there a test for profile_completion auto-completion (bio + avatar)? [Coverage, quickstart.md Scenario 2]
^- [X] CHK010 - Is there a test for social_links auto-completion? [Coverage, quickstart.md Scenario 9]
^- [X] CHK011 - Is there a test that partial data does NOT trigger auto-completion? [Coverage, Gap]
^- [X] CHK012 - Is there a test that step completion is permanent (no reversal)? [Coverage, quickstart.md Scenario 10]

## Authorization Tests

^- [X] CHK013 - Is there a test for non-owner access denied (401)? [Coverage, quickstart.md Scenario 7]
^- [X] CHK014 - Is there a test for admin activation endpoint? [Coverage, quickstart.md Scenario 8]
^- [X] CHK015 - Is there a test for required step cannot be skipped (403)? [Coverage, quickstart.md Scenario 3]

## Next-Step Logic Tests

^- [X] CHK016 - Is there a test for next-step returns correct first incomplete step? [Coverage, quickstart.md Scenario 5]
^- [X] CHK017 - Is there a test for next-step returns null when all steps completed? [Coverage, quickstart.md Scenario 5]
^- [X] CHK018 - Is there a test for recalculate adds missing steps? [Coverage, quickstart.md Scenario 6]

## Recalculate Tests

^- [X] CHK019 - Is there a test that recalculate adds new steps from template update? [Coverage, quickstart.md Scenario 6]
^- [X] CHK020 - Is there a test that recalculate preserves existing completed steps? [Coverage, quickstart.md Scenario 6]

## Role-Specific Tests

^- [X] CHK021 - Is there a test for Brand profile activation flow? [Coverage, quickstart.md Scenario 8]
^- [X] CHK022 - Is there a test for Editor profile activation flow? [Coverage, quickstart.md Scenario 1]
^- [X] CHK023 - Is there a test for Influencer profile with social links? [Coverage, quickstart.md Scenario 9]

## Notes

Test coverage should address all critical paths defined in the quickstart scenarios and edge cases.