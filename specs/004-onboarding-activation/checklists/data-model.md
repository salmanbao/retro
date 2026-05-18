# Data Model Integrity Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate entity definitions, relationships, and data integrity rules
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md), [data-model.md](../data-model.md)

## Entity Definitions

^- [X] CHK001 - Are all four entities defined (OnboardingTemplate, OnboardingStep, OnboardingProgress, StepProgress)? [Completeness, Data Model §Entity Overview]
^- [X] CHK002 - Does OnboardingTemplate include all required fields (id, profile_type, version, timestamps)? [Completeness, Data Model §OnboardingTemplate]
^- [X] CHK003 - Does OnboardingStep include all required fields (id, template_id, title, step_type, required, display_order)? [Completeness, Data Model §OnboardingStep]
^- [X] CHK004 - Does OnboardingProgress include all required fields (id, profile_id, template_id, template_version, activation_status)? [Completeness, Data Model §OnboardingProgress]
^- [X] CHK005 - Does StepProgress include all required fields (id, onboarding_progress_id, step_id, status, timestamps)? [Completeness, Data Model §StepProgress]

## Relationships

^- [X] CHK006 - Is the 1:N relationship between OnboardingTemplate and OnboardingStep documented? [Completeness, Data Model §Relationships]
^- [X] CHK007 - Is the 1:N relationship between OnboardingTemplate and OnboardingProgress documented? [Completeness, Data Model §Relationships]
^- [X] CHK008 - Is the 1:1 relationship between Profile and OnboardingProgress documented? [Completeness, Data Model §Relationships]
^- [X] CHK009 - Is the 1:N relationship between OnboardingProgress and StepProgress documented? [Completeness, Data Model §Relationships]

## Unique Constraints

^- [X] CHK010 - Is the unique constraint on (profile_type, version) for OnboardingTemplate specified? [Completeness, Data Model §OnboardingTemplate]
^- [X] CHK011 - Is the unique constraint on profile_id for OnboardingProgress specified? [Completeness, Data Model §OnboardingProgress]

## Indexes

^- [X] CHK012 - Is the index on template_id for OnboardingStep specified? [Completeness, Data Model §OnboardingStep]
^- [X] CHK013 - Is the index on profile_id (unique) for OnboardingProgress specified? [Completeness, Data Model §OnboardingProgress]
^- [X] CHK014 - Is the index on activation_status for OnboardingProgress specified? [Completeness, Data Model §OnboardingProgress]
^- [X] CHK015 - Are indexes on onboarding_progress_id and step_id for StepProgress specified? [Completeness, Data Model §StepProgress]

## State Transition Fields

^- [X] CHK016 - Are started_at and completed_at timestamps tracked for step progress? [Completeness, Data Model §StepProgress]
^- [X] CHK017 - Are started_at and last_activity_at timestamps tracked for onboarding progress? [Completeness, Data Model §OnboardingProgress]

## Notes

Data model integrity ensures referential consistency and query performance.