package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// AuthzHandler handles authorization-related HTTP endpoints.
type AuthzHandler struct {
	authzSvc *service.AuthorizationService
}

// NewAuthzHandler creates a new AuthzHandler.
func NewAuthzHandler(authzSvc *service.AuthorizationService) *AuthzHandler {
	return &AuthzHandler{
		authzSvc: authzSvc,
	}
}

// RegisterRoutes registers authorization routes on the chi router.
func (h *AuthzHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/roles", h.ListRoles)
	r.Get("/api/v1/permissions", h.ListPermissions)
	r.Post("/api/v1/profiles/{profileID}/roles", h.AssignRole)
	r.Delete("/api/v1/profiles/{profileID}/roles/{roleID}", h.RevokeRole)
	r.Get("/api/v1/profiles/{profileID}/permissions", h.GetEffectivePermissions)
}

// ListRolesResponse represents the response for listing roles.
type ListRolesResponse struct {
	Roles []RoleResponse `json:"roles"`
}

// RoleResponse represents a role in API responses.
type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ParentID    *string  `json:"parent_id,omitempty"`
	Permissions []string `json:"permissions"`
}

// ListRoles handles GET /api/v1/roles.
func (h *AuthzHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.authzSvc.ListRoles(r.Context())
	if err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to list roles")
		return
	}

	response := ListRolesResponse{
		Roles: make([]RoleResponse, len(roles)),
	}
	for i, role := range roles {
		var parentID *string
		if role.ParentID != nil {
			pid := role.ParentID.String()
			parentID = &pid
		}

		// Get permissions for this role
		perms, err := h.authzSvc.GetRolePermissions(r.Context(), role.ID)
		if err != nil {
			perms = []string{}
		}

		response.Roles[i] = RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			ParentID:    parentID,
			Permissions: perms,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// ListPermissionsResponse represents the response for listing permissions.
type ListPermissionsResponse struct {
	Permissions []PermissionResponse `json:"permissions"`
}

// PermissionResponse represents a permission in API responses.
type PermissionResponse struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
}

// ListPermissions handles GET /api/v1/permissions.
func (h *AuthzHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.authzSvc.ListPermissions(r.Context())
	if err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to list permissions")
		return
	}

	response := ListPermissionsResponse{
		Permissions: make([]PermissionResponse, len(perms)),
	}
	for i, p := range perms {
		response.Permissions[i] = PermissionResponse{
			Key:         p.Key,
			Description: p.Description,
			Domain:      string(p.Domain),
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// AssignRoleRequest represents the request body for assigning a role.
type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}

// AssignRoleResponse represents the response for role assignment.
type AssignRoleResponse struct {
	ProfileID  string `json:"profile_id"`
	RoleID     string `json:"role_id"`
	AssignedAt string `json:"assigned_at"`
}

// AssignRole handles POST /api/v1/profiles/{profileID}/roles.
func (h *AuthzHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	profileIDStr := chi.URLParam(r, "profileID")
	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid profile ID")
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid role ID")
		return
	}

	if err := h.authzSvc.AssignRole(r.Context(), profileID, roleID); err != nil {
		if err.Error() == "maximum roles per profile exceeded" {
			writeAuthzError(w, http.StatusUnprocessableEntity, "unprocessable", err.Error())
			return
		}
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to assign role")
		return
	}

	response := AssignRoleResponse{
		ProfileID: profileID.String(),
		RoleID:    roleID.String(),
	}

	writeJSON(w, http.StatusCreated, response)
}

// RevokeRole handles DELETE /api/v1/profiles/{profileID}/roles/{roleID}.
func (h *AuthzHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	profileIDStr := chi.URLParam(r, "profileID")
	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid profile ID")
		return
	}

	roleIDStr := chi.URLParam(r, "roleID")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid role ID")
		return
	}

	if err := h.authzSvc.RevokeRole(r.Context(), profileID, roleID); err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to revoke role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEffectivePermissionsResponse represents the response for getting effective permissions.
type GetEffectivePermissionsResponse struct {
	ProfileID   string   `json:"profile_id"`
	Permissions []string `json:"permissions"`
	Roles       []RoleInfo `json:"roles"`
}

// RoleInfo represents role information in the effective permissions response.
type RoleInfo struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	InheritedFrom *string `json:"inherited_from,omitempty"`
}

// GetEffectivePermissions handles GET /api/v1/profiles/{profileID}/permissions.
// Requires caller to have profile.view permission (FR-AUTH-010).
func (h *AuthzHandler) GetEffectivePermissions(w http.ResponseWriter, r *http.Request) {
	profileIDStr := chi.URLParam(r, "profileID")
	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		writeAuthzError(w, http.StatusBadRequest, "bad_request", "invalid profile ID")
		return
	}

	// Check if caller has profile.view permission (FR-AUTH-010)
	callerProfileID := middleware.GetCallerProfileID(r.Context())
	if callerProfileID == nil {
		writeAuthzError(w, http.StatusForbidden, "forbidden", "no active profile context")
		return
	}

	hasViewPermission, err := h.authzSvc.HasPermission(r.Context(), *callerProfileID, "profile.view")
	if err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to check permissions")
		return
	}
	if !hasViewPermission {
		writeAuthzError(w, http.StatusForbidden, "forbidden", "permission denied")
		return
	}

	perms, err := h.authzSvc.GetEffectivePermissions(r.Context(), profileID)
	if err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to get effective permissions")
		return
	}

	// Get profile roles with inheritance info
	profileRoles, err := h.authzSvc.GetProfileRoles(r.Context(), profileID)
	if err != nil {
		writeAuthzError(w, http.StatusInternalServerError, "internal_error", "failed to get profile roles")
		return
	}

	roles := make([]RoleInfo, 0, len(profileRoles))
	for _, pr := range profileRoles {
		role, err := h.authzSvc.GetRoleByID(r.Context(), pr.RoleID)
		if err != nil {
			continue
		}

		roleInfo := RoleInfo{
			ID:   role.ID.String(),
			Name: role.Name,
		}

		// Check if this role was inherited (has a parent role)
		if role.ParentID != nil {
			parentRole, err := h.authzSvc.GetRoleByID(r.Context(), *role.ParentID)
			if err == nil {
				roleInfo.InheritedFrom = &parentRole.Name
			}
		}

		roles = append(roles, roleInfo)
	}

	response := GetEffectivePermissionsResponse{
		ProfileID:   profileID.String(),
		Permissions: perms,
		Roles:       roles,
	}

	writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeAuthzError writes an error response.
func writeAuthzError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errorCode,
		"message": message,
	})
}