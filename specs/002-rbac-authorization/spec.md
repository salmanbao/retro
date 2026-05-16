# Feature Specification: Authorization Module

**Feature Branch**: `[002-rbac-authorization]`

**Created**: 2026-05-15

**Status**: Draft

**Input**: User description: "Build the Authorization module for ViralForge. The module manages role-based access control (RBAC), permission policies, and role hierarchies for Brand, Editor, Influencer, and platform profiles.

Authorization is attached to profiles rather than users, allowing one user account to hold different roles across multiple profiles."

## Clarifications

### Session 2026-05-15

- Q: Role deletion cascade behavior? → A: Cascade delete - Remove all ProfileRole assignments for that role automatically
- Q: Who can assign roles to profiles? → A: Platform admin only - Only platform_admin role can assign/remove roles
- Q: Invalid permission string handling? → A: Return false - hasPermission returns false for non-existent permissions; middleware allows (passes through)

## User Scenarios & Testing *(mandatory)*

### Scenario 1 - Brand Profile Authorization

A Brand profile owner needs to control what team members can do. The brand_owner role has full permissions, while brand_marketer has limited permissions for campaign content only.

**Independent Test**: Assign roles to a Brand profile, verify permission resolution includes inherited permissions from role hierarchy.

---

### Scenario 2 - Editor Permission Scope

An Editor profile should only access submission review features. The editor_senior role grants submission.review but not campaign management.

**Independent Test**: Create Editor profile with editor_senior role, call hasPermission(profileID, "submission.review") returns true, hasPermission(profileID, "campaign.create") returns false.

---

### Scenario 3 - Influencer Role Assignment

An Influencer profile owner assigns influencer_owner role to their profile. The role grants analytics.view and offer.claim permissions but not campaign management.

**Independent Test**: Assign influencer_owner role to Influencer profile, verify effective permissions include analytics.view and offer.claim.

---

### Scenario 4 - Platform Admin Global Access

A platform_admin role has global permissions across all domains. When assigned to a Platform profile, the profile inherits all permissions.

**Independent Test**: Assign platform_admin to Platform profile, verify hasPermission returns true for permissions from multiple domains.

---

### Scenario 5 - Role Hierarchy Inheritance

A brand_admin role inherits from brand_owner. When a profile has brand_admin, it should have both brand_admin direct permissions and inherited brand_owner permissions.

**Independent Test**: Create role hierarchy brand_admin -> brand_owner, assign brand_admin to Brand profile, verify effective permissions include both brand_admin's direct permissions AND brand_owner's permissions.

---

### Scenario 6 - Authorization Middleware Enforcement

An API endpoint requires campaign.create permission. A request with a Brand profile that has brand_marketer role (without campaign.create) should be denied.

**Independent Test**: Request to protected endpoint with profile lacking required permission returns 403 Forbidden.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-001**: The system MUST define permissions as unique strings in dot-notation format (e.g., "resource.action").
- **FR-AUTH-002**: The system MUST define roles as named entities with associated permissions.
- **FR-AUTH-003**: The system MUST support role hierarchies where child roles inherit permissions from parent roles.
- **FR-AUTH-004**: The system MUST assign roles to profiles (many roles per profile allowed).
- **FR-AUTH-005**: The system MUST resolve effective permissions for a profile by aggregating direct role permissions and inherited permissions from role hierarchy.
- **FR-AUTH-006**: The system MUST provide hasPermission(profileID, permission) function returning boolean.
- **FR-AUTH-007**: The system MUST provide requirePermission(permission) middleware that rejects requests without the permission.
- **FR-AUTH-008**: The system MUST expose REST API for role management: list roles, assign role to profile, remove role from profile.
- **FR-AUTH-009**: The system MUST expose REST API to list available permissions.
- **FR-AUTH-010**: The system MUST expose REST API to get effective permissions for a profile.
- **FR-AUTH-011**: The system MUST record created_at and updated_at timestamps on role assignments.
- **FR-AUTH-012**: Automated tests MUST cover domain entities, permission inheritance, policy evaluation, repositories, and HTTP handlers.
- **FR-AUTH-013**: The system MUST return false for hasPermission calls with non-existent permission strings (no error).

### Supported Profile Domains

- **Brand**: Campaign management, offer management
- **Editor**: Submission review, content validation
- **Influencer**: Offer claiming, analytics viewing
- **Platform**: System-wide administration

### Defined Permissions

| Permission | Description | Domain |
|------------|-------------|--------|
| campaign.create | Create new campaigns | Brand |
| campaign.update | Modify existing campaigns | Brand |
| campaign.delete | Delete campaigns | Brand |
| campaign.view | View campaign details | Brand, Editor, Influencer |
| submission.review | Review submissions | Editor |
| submission.approve | Approve submissions | Editor |
| submission.reject | Reject submissions | Editor |
| offer.claim | Claim offers | Influencer |
| offer.view | View available offers | Influencer |
| wallet.view | View wallet balance | Influencer |
| analytics.view | View performance analytics | Influencer, Brand |
| profile.manage | Manage profile settings | All |
| role.assign | Assign roles to profiles | Platform |
| role.revoke | Remove roles from profiles | Platform |
| role.create | Create new roles | Platform |
| role.delete | Delete roles | Platform |
| permission.assign | Assign permissions to roles | Platform |
| permission.revoke | Remove permissions from roles | Platform |

### Defined Roles

| Role | Description | Permissions | Parent Role |
|------|-------------|--------------|-------------|
| brand_owner | Full Brand access | campaign.*, analytics.view, offer.view, profile.manage | - |
| brand_admin | Brand administration | campaign.create, campaign.update, analytics.view, profile.manage | brand_owner |
| brand_marketer | Brand marketing | campaign.create, campaign.view, profile.manage | brand_admin |
| editor_owner | Full Editor access | submission.*, profile.manage | - |
| editor_senior | Senior Editor | submission.review, submission.approve, profile.manage | editor_owner |
| editor_junior | Junior Editor | submission.review, profile.manage | editor_senior |
| influencer_owner | Full Influencer access | offer.*, wallet.view, analytics.view, profile.manage | - |
| influencer_partner | Influencer partner | offer.claim, offer.view, profile.manage | influencer_owner |
| platform_admin | Platform administration | * (all permissions) | - |

### Role Hierarchy Rules

1. A role inherits ALL permissions from its parent role.
2. Multiple inheritance is NOT supported (single parent only).
3. Role hierarchy depth maximum: 3 levels.
4. Circular inheritance is prevented.
5. Removing a parent role does not remove inherited permissions already assigned to child role's permissions (they remain as direct assignments).
6. Deleting a role cascade-deletes all ProfileRole assignments for that role.
7. Only platform_admin can assign or remove roles from profiles (role.assign and role.revoke permissions required).

## Key Entities *(include if feature involves data)*

- **Permission**: Represents a system capability. Contains permission key (unique string), description, and domain.
- **Role**: Represents a named collection of permissions. Contains name, description, optional parent role reference, and timestamps.
- **RolePermission**: Join table mapping roles to permissions (many-to-many).
- **ProfileRole**: Join table mapping profiles to roles (many-to-many) with timestamps.
- **RoleHierarchy**: Represents parent-child relationship between roles for inheritance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-001**: Permission resolution completes within 50ms for profiles with up to 5 roles.
- **SC-AUTH-002**: Authorization middleware adds less than 10ms latency to protected endpoints.
- **SC-AUTH-003**: Role assignment or removal takes effect within 100ms.
- **SC-AUTH-004**: System correctly resolves inherited permissions through up to 3-level role hierarchy.
- **SC-AUTH-005**: All automated tests pass for domain logic, application services, repositories, and API endpoints.
- **SC-AUTH-006**: System handles concurrent authorization requests with no permission leakage between requests.

## Assumptions

- Profile authorization is independent of user authentication (authentication handled separately).
- Roles and permissions are **seeded** at system initialization (19 permissions, 9 roles as defined in this specification). Platform admins with role.create/role.delete permissions CAN dynamically create or remove roles via the REST API.
- Maximum roles per profile: 10.
- Maximum permissions per role: 50.
- Role hierarchy is acyclic (no circular dependencies).
- Permissions follow dot-notation: "resource.action" format.
- Wildcard permissions (e.g., "campaign.*") grant all actions on the resource.