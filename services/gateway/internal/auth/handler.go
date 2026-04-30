package auth

import (
	"encoding/json"
	"net/http"

	authpb "github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/pb/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req authpb.LoginRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	token, err := h.service.Authenticate(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req authpb.RegisterRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	token, err := h.service.Register(r.Context(), &req)
	if err != nil {
		// The auth service returns ErrUserAlredyExists only when password doesn't match.
		// A matching password returns a token (idempotent register).
		http.Error(w, `{"error":"User already exists"}`, http.StatusForbidden)
		return
	}

	// Success: either newly created (201) or idempotent match (200).
	// We use 200 for simplicity — the caller only cares about the token.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}
