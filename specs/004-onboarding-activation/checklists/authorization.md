# Authorization Checklist: Role-Based Onboarding and Activation

**Purpose**: Validate ownership and authorization boundary clarity
**Created**: 2026-05-18
**Feature**: [spec.md](../spec.md)

## Ownership Verification

^- [X] CHK001 - Is profile ownership verification required for all onboarding endpoints? [Completeness, Spec §BR-002]
^- [X] CHK002 - Is the ownership check applied to GET /onboarding? [Coverage, Spec §BR-002]
^- [X] CHK003 - Is the ownership check applied to GET /onboarding/steps? [Coverage, Spec §BR-002]
^- [X] CHK004 - Is the ownership check applied to PATCH /onboarding/steps/{stepId}? [Coverage, Spec §BR-002]
^- [X] CHK005 - Is the ownership check applied to POST /onboarding/recalculate? [Coverage, Spec §BR-002]
^- [X] CHK006 - Is the ownership check applied to GET /onboarding/next-step? [Coverage, Spec §BR-002]

## Admin Authorization

^- [X] CHK007 - Is admin role verification required for POST /admin/profiles/{id}/onboarding/activate? [Completeness, Spec §BR-005]
^- [X] CHK008 - Is the admin endpoint documented in API endpoints section? [Completeness, Spec §4]
^- [X] CHK009 - Are the authorization requirements consistent between spec and API docs? [Consistency, Spec §4]

## Cross-Profile Access

^- [X] CHK010 - Is cross-profile access (user B accessing user A's onboarding) explicitly blocked? [Completeness, Gap]
^- [X] CHK011 - Is the 401 response for unauthorized access documented? [Completeness, Spec §API Errors]
^- [X] CHK012 - Are there any scenarios where non-owner access is permitted? [Ambiguity, Gap]

## Multi-Profile Users

^- [X] CHK013 - Is the scenario of a user with multiple profiles (Brand + Editor) addressed? [Completeness, Spec §BR-001]
^- [X] CHK014 - Is each profile's onboarding tracked independently documented? [Clarity, Spec §BR-001]
^- [X] CHK015 - Does ownership of one profile grant access to other profiles? [Ambiguity, Gap]

## Notes

Authorization boundaries prevent users from accessing or modifying other users' onboarding progress.