package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type SearchHandler struct {
	repo Repository
}

func NewSearchHandler(repo Repository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

type UserResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
}

func (h *SearchHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 50 {
		limit = 50
	}

	ctx := context.Background()
	users, err := h.repo.SearchByUsername(ctx, query, limit)
	if err != nil {
		http.Error(w, `{"error":"Failed to search users"}`, http.StatusInternalServerError)
		return
	}

	// Convert to response format (exclude passwords)
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			// NOTE: DeviceTokens intentionally omitted from search results (public-facing)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type GetUsersRequest struct {
	IDs []string `json:"ids"`
}

func (h *SearchHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var req GetUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, `{"error":"ids array is required"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	users, err := h.repo.GetUsersByIDs(ctx, req.IDs)
	if err != nil {
		http.Error(w, `{"error":"Failed to get users"}`, http.StatusInternalServerError)
		return
	}

	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// internalUserResponse includes device tokens — only for service-to-service calls.
type internalUserResponse struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name,omitempty"`
	DeviceTokens []string `json:"device_tokens,omitempty"`
}

// GetUsersInternal is the service-to-service variant of GetUsers.
// It returns device tokens and is only exposed on the internal HTTP port (not via the gateway).
func (h *SearchHandler) GetUsersInternal(w http.ResponseWriter, r *http.Request) {
	var req GetUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, `{"error":"ids array is required"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	users, err := h.repo.GetUsersByIDs(ctx, req.IDs)
	if err != nil {
		http.Error(w, `{"error":"Failed to get users"}`, http.StatusInternalServerError)
		return
	}

	response := make([]internalUserResponse, len(users))
	for i, user := range users {
		response[i] = internalUserResponse{
			ID:           user.ID,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			DeviceTokens: user.DeviceTokens,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type DeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (h *SearchHandler) AddDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req DeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, `{"error":"Invalid request or missing token"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.repo.AddDeviceToken(ctx, userID, req.Token); err != nil {
		http.Error(w, `{"error":"Failed to save token"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (h *SearchHandler) RemoveDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req DeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, `{"error":"Invalid request or missing token"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.repo.RemoveDeviceToken(ctx, userID, req.Token); err != nil {
		http.Error(w, `{"error":"Failed to remove token"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

func (h *SearchHandler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	// Parse user_id from headers (set by gateway after JWT auth)
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req UpdateDisplayNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize: trim whitespace and enforce max length
	displayName := strings.TrimSpace(req.DisplayName)
	if len(displayName) > 100 {
		displayName = displayName[:100]
	}

	ctx := context.Background()
	if err := h.repo.UpdateDisplayName(ctx, userID, displayName); err != nil {
		http.Error(w, `{"error":"Failed to update display name"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
