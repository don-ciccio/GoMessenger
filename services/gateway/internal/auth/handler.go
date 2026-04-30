package auth

import (
	"encoding/json"
	"net/http"
	"os"

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

	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Secret   string `json:"secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
		return
	}

	// Security: Prevent random users from registering manually
	// The Shopify app will pass this secret when provisioning merchants.
	expectedSecret := os.Getenv("REGISTRATION_SECRET")
	if expectedSecret != "" && reqData.Secret != expectedSecret {
		http.Error(w, `{"error":"Registration is restricted"}`, http.StatusForbidden)
		return
	}

	req := authpb.RegisterRequest{
		Username: reqData.Username,
		Password: reqData.Password,
	}

	token, err := h.service.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"User already exists"}`, http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}
