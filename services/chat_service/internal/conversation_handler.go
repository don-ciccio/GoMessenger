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
	conversation, err := h.conversationService.GetOrCreateConversation(ctx, req.Participants, req.ShopID)
	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// ListConversations handles GET /conversations?user_id=xxx&shop_id=yyy
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Authorization: ensure the caller is requesting their own conversations
	authenticatedUserID := r.Header.Get("X-User-Id")
	if authenticatedUserID != "" && authenticatedUserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	shopID := r.URL.Query().Get("shop_id")

	ctx := r.Context()
	conversations, err := h.conversationService.ListUserConversations(ctx, userID, shopID)
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

	// Authorization: verify the caller is a participant in this conversation
	authenticatedUserID := r.Header.Get("X-User-Id")
	if authenticatedUserID != "" {
		if err := h.conversationService.ValidateUserInConversation(r.Context(), conversationID, authenticatedUserID); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
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

	// Reverse messages to return them in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// ArchiveConversation handles POST /conversations/{id}/archive
func (h *ConversationHandler) ArchiveConversation(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	if conversationID == "" {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.conversationService.ArchiveConversation(r.Context(), conversationID, userID); err != nil {
		http.Error(w, "Failed to archive conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "archived"})
}

// UnarchiveConversation handles POST /conversations/{id}/unarchive
func (h *ConversationHandler) UnarchiveConversation(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	if conversationID == "" {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.conversationService.UnarchiveConversation(r.Context(), conversationID, userID); err != nil {
		http.Error(w, "Failed to unarchive conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unarchived"})
}

// ListArchivedConversations handles GET /conversations/archived?user_id=xxx
func (h *ConversationHandler) ListArchivedConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	authenticatedUserID := r.Header.Get("X-User-Id")
	if authenticatedUserID != "" && authenticatedUserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	conversations, err := h.conversationService.ListArchivedConversations(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to list archived conversations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}
