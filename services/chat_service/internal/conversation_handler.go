package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type ConversationHandler struct {
	conversationService *ConversationService
	messageRepo         Repository
}

func NewConversationHandler(conversationService *ConversationService, messageRepo Repository) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		messageRepo:         messageRepo,
	}
}

// CreateOrGetConversation handles POST /conversations
func (h *ConversationHandler) CreateOrGetConversation(w http.ResponseWriter, r *http.Request) {
	var req ConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	conversation, err := h.conversationService.GetOrCreateConversation(ctx, req.Participants)
	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// ListConversations handles GET /conversations?user_id=xxx
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	conversations, err := h.conversationService.ListUserConversations(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to list conversations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// GetConversationMessages handles GET /conversations/:id/messages?limit=50&offset=0
func (h *ConversationHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	if conversationID == "" {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	ctx := context.Background()
	messages, err := h.messageRepo.FindByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
