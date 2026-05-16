# Contracts: Authorization Module

**Feature**: 002-rbac-authorization | **Date**: 2026-05-15

## REST API Endpoints

### GET /api/v1/roles

List all available roles.

**Request**: None

**Response 200**:
```json
{
  "roles": [
    {
      "id": "uuid",
      "name": "brand_owner",
      "description": "Full Brand access",
      "parent_id": null,
      "permissions": ["campaign.*", "analytics.view", "offer.view", "profile.manage"],
      "created_at": "2026-05-15T10:00:00Z",
      "updated_at": "2026-05-15T10:00:00Z"
    }
  ]
}
```

---

### GET /api/v1/permissions

List all available permissions.

**Request**: None

**Response 200**:
```json
{
  "permissions": [
    {
      "key": "campaign.create",
      "description": "Create new campaigns",
      "domain": "Brand"
    }
  ]
}
```

---

### POST /api/v1/profiles/{id}/roles

Assign a role to a profile.

**Request**:
```json
{
  "role_id": "uuid"
}
```

**Response 201**:
```json
{
  "profile_id": "uuid",
  "role_id": "uuid",
  "assigned_at": "2026-05-15T10:00:00Z"
}
```

**Errors**:
- 400: Invalid request body
- 403: Caller lacks role.assign permission
- 404: Profile or role not found
- 409: Role already assigned to profile
- 422: Max roles per profile exceeded

---

### DELETE /api/v1/profiles/{id}/roles/{roleId}

Remove a role from a profile.

**Request**: None

**Response 204**: No content

**Errors**:
- 403: Caller lacks role.revoke permission
- 404: Profile or role not found
- 409: Role not assigned to profile

---

### GET /api/v1/profiles/{id}/permissions

Get effective permissions for a profile.

**Request**: None

**Response 200**:
```json
{
  "profile_id": "uuid",
  "permissions": [
    "campaign.create",
    "campaign.update",
    "campaign.view",
    "analytics.view",
    "profile.manage"
  ],
  "roles": [
    {
      "id": "uuid",
      "name": "brand_admin",
      "inherited_from": "brand_owner"
    }
  ]
}
```

**Errors**:
- 403: Caller lacks permission to view
- 404: Profile not found

---

## Middleware Contract

### requirePermission(permission string) Middleware

**Behavior**:
1. Extract profile ID from request context (set by AuthMiddleware)
2. Call `hasPermission(profileID, permission)`
3. If false → Return 403 Forbidden with JSON body:
   ```json
   {
     "error": "forbidden",
     "message": "permission denied: {permission}"
   }
   ```
4. If true → Pass request to next handler

**Integration**:
```go
r.Route("/campaigns", func(r chi.Router) {
    r.Use(AuthMiddleware.Authenticate)
    r.With(requirePermission("campaign.create")).Post("/", h.Create)
})
```

---

## Error Responses

All errors follow this format:
```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

| HTTP Code | error | Description |
|-----------|-------|-------------|
| 400 | bad_request | Invalid request format |
| 403 | forbidden | Permission denied |
| 404 | not_found | Resource not found |
| 409 | conflict | State conflict |
| 422 | unprocessable | Validation failed |
| 500 | internal_error | Server error |