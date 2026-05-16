package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// PortfolioItemResponse represents a portfolio item in API responses.
type PortfolioItemResponse struct {
	ID           string `json:"id"`
	ProfileID    string `json:"profile_id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	VideoURL     string `json:"video_url,omitempty"`
	ExternalLink string `json:"external_link,omitempty"`
	DisplayOrder int    `json:"display_order"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreatePortfolioItemRequest represents POST /api/v1/profiles/{id}/portfolio request.
type CreatePortfolioItemRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	VideoURL     string `json:"video_url,omitempty"`
	ExternalLink string `json:"external_link,omitempty"`
	DisplayOrder int    `json:"display_order"`
}

// UpdatePortfolioItemRequest represents PATCH /api/v1/profiles/{id}/portfolio/{itemId} request.
type UpdatePortfolioItemRequest struct {
	Title        *string `json:"title,omitempty"`
	Description  *string `json:"description,omitempty"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	VideoURL     *string `json:"video_url,omitempty"`
	ExternalLink *string `json:"external_link,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty"`
}

// PortfolioHandler handles portfolio CRUD HTTP endpoints.
type PortfolioHandler struct {
	portfolioSvc *service.PortfolioService
}

// NewPortfolioHandler creates a new PortfolioHandler.
func NewPortfolioHandler(portfolioSvc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{portfolioSvc: portfolioSvc}
}

// GetPortfolio handles GET /api/v1/profiles/{id}/portfolio.
func (h *PortfolioHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	items, err := h.portfolioSvc.ListByProfileID(r.Context(), profileID, 50, 0)
	if err != nil {
		if err == service.ErrProfileNotEditor {
			writeError(w, http.StatusForbidden, "forbidden", "Portfolio items are only available for Editor profiles")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve portfolio items")
		return
	}

	response := make([]PortfolioItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toPortfolioItemResponse(item))
	}

	writeJSON(w, http.StatusOK, response)
}

// CreatePortfolioItem handles POST /api/v1/profiles/{id}/portfolio.
func (h *PortfolioHandler) CreatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req CreatePortfolioItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}

	item, err := h.portfolioSvc.Create(r.Context(), profileID, req.Title, req.Description, req.ThumbnailURL, req.VideoURL, req.ExternalLink, req.DisplayOrder)
	if err != nil {
		if err == service.ErrProfileNotEditor {
			writeError(w, http.StatusForbidden, "forbidden", "Portfolio items are only available for Editor profiles")
			return
		}
		if err == service.ErrPortfolioLimitReached {
			writeError(w, http.StatusForbidden, "limit_reached", "Maximum portfolio items (50) reached for this profile")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create portfolio item")
		return
	}

	writeJSON(w, http.StatusCreated, toPortfolioItemResponse(item))
}

// UpdatePortfolioItem handles PATCH /api/v1/profiles/{id}/portfolio/{itemId}.
func (h *PortfolioHandler) UpdatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid item ID format")
		return
	}

	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req UpdatePortfolioItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Build update values with defaults
	title := ""
	description := ""
	thumbnailURL := ""
	videoURL := ""
	externalLink := ""
	displayOrder := 0

	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.ThumbnailURL != nil {
		thumbnailURL = *req.ThumbnailURL
	}
	if req.VideoURL != nil {
		videoURL = *req.VideoURL
	}
	if req.ExternalLink != nil {
		externalLink = *req.ExternalLink
	}
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	}

	item, err := h.portfolioSvc.Update(r.Context(), itemID, profileID, title, description, thumbnailURL, videoURL, externalLink, displayOrder)
	if err != nil {
		if err == service.ErrProfileNotEditor {
			writeError(w, http.StatusForbidden, "forbidden", "Portfolio items are only available for Editor profiles")
			return
		}
		if err == domain.ErrPortfolioItemNotFound {
			writeError(w, http.StatusNotFound, "not_found", "Portfolio item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update portfolio item")
		return
	}

	writeJSON(w, http.StatusOK, toPortfolioItemResponse(item))
}

// DeletePortfolioItem handles DELETE /api/v1/profiles/{id}/portfolio/{itemId}.
func (h *PortfolioHandler) DeletePortfolioItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid item ID format")
		return
	}

	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	err = h.portfolioSvc.Delete(r.Context(), itemID, profileID)
	if err != nil {
		if err == service.ErrProfileNotEditor {
			writeError(w, http.StatusForbidden, "forbidden", "Portfolio items are only available for Editor profiles")
			return
		}
		if err == domain.ErrPortfolioItemNotFound {
			writeError(w, http.StatusNotFound, "not_found", "Portfolio item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete portfolio item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toPortfolioItemResponse converts a domain PortfolioItem to a response struct.
func toPortfolioItemResponse(item *domain.PortfolioItem) PortfolioItemResponse {
	return PortfolioItemResponse{
		ID:           item.ID.String(),
		ProfileID:    item.ProfileID.String(),
		Title:        item.Title,
		Description:  item.Description,
		ThumbnailURL: item.ThumbnailURL,
		VideoURL:     item.VideoURL,
		ExternalLink: item.ExternalLink,
		DisplayOrder: item.DisplayOrder,
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
