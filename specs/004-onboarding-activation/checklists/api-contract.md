# API Contract Quality Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate API endpoint definitions, request/response contracts, and error handling
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md), [quickstart.md](../quickstart.md)

## Endpoint Coverage

^- [X] CHK001 - Is GET /api/v1/profiles/{id}/onboarding defined with response schema? [Completeness, Spec §4]
^- [X] CHK002 - Is GET /api/v1/profiles/{id}/onboarding/steps defined with response schema? [Completeness, Spec §4]
^- [X] CHK003 - Is PATCH /api/v1/profiles/{id}/onboarding/steps/{stepId} defined with request/response schema? [Completeness, Spec §4]
^- [X] CHK004 - Is POST /api/v1/profiles/{id}/onboarding/recalculate defined with response schema? [Completeness, Spec §4]
^- [X] CHK005 - Is GET /api/v1/profiles/{id}/onboarding/next-step defined with response schema? [Completeness, Spec §4]
^- [X] CHK006 - Is POST /api/v1/admin/profiles/{id}/onboarding/activate defined with response schema? [Completeness, Spec §4]

## Request Validation

^- [X] CHK007 - Is the PATCH request body format (status field only) specified? [Clarity, Spec §4]
^- [X] CHK008 - Are valid status values enumerated for PATCH requests? [Completeness, Spec §4]
^- [X] CHK009 - Are invalid status transitions explicitly defined as 400 errors? [Completeness, Spec §Errors]

## Response Schemas

^- [X] CHK010 - Does GET /onboarding response include profile_id, activation_status, percentage_complete, required_steps_remaining, template_version, timestamps? [Completeness, Spec §GET /onboarding Response]
^- [X] CHK011 - Does GET /onboarding/steps response include steps array with id, title, description, action_url, step_type, required, display_order, status, timestamps? [Completeness, Spec §GET /onboarding/steps Response]
^- [X] CHK012 - Does GET /onboarding/next-step response include step object or null with message? [Completeness, Spec §GET /onboarding/next-step Response]

## Error Responses

^- [X] CHK013 - Is 401 Unauthorized documented for non-owner access? [Completeness, Spec §Errors]
^- [X] CHK014 - Is 403 Forbidden documented for required step skip attempts? [Completeness, Spec §Errors]
^- [X] CHK015 - Is 404 Not Found documented for missing profile or step? [Completeness, Spec §Errors]
^- [X] CHK016 - Is 400 Bad Request documented for invalid state transitions? [Completeness, Spec §Errors]

## Admin Endpoint

^- [X] CHK017 - Is the admin activate endpoint (POST /admin/profiles/{id}/onboarding/activate) documented with proper authorization? [Completeness, Spec §4]
^- [X] CHK018 - Is 400 Bad Request for invalid activation state (not pending_review) documented? [Completeness, Spec §Errors]

## Contract Testability

^- [X] CHK019 - Are all API contracts testable with clear pass/fail criteria? [Measurability, Spec §4]
^- [X] CHK020 - Does the quickstart.md include contract test scenarios? [Completeness, Spec §quickstart.md]

## Notes

API contracts define the interface between client and server; all contracts must be testable.