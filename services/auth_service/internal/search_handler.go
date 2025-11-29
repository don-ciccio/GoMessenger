package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type SearchHandler struct {
	repo Repository
}

func NewSearchHandler(repo Repository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
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
			ID:       user.ID,
			Username: user.Username,
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
			ID:       user.ID,
			Username: user.Username,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
