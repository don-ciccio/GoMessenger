package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u LoginUserRequest
	json.NewDecoder(r.Body).Decode(&u)
	fmt.Printf("The user request value %v", u)

	token, err := h.service.Authenticate(u)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	resp := AuthResponse{Token: token}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u RegisterUserRequest
	json.NewDecoder(r.Body).Decode(&u)
	fmt.Printf("The user request value %v", u)

	token, err := h.service.Register(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error registering user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := AuthResponse{Token: token}
	json.NewEncoder(w).Encode(resp)
}
