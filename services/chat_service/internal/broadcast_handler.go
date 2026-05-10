package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type BroadcastHandler struct {
	broadcastRepo       BroadcastRepository
	conversationService *ConversationService
	rdb                 *redis.Client
	streamName          string
}

func NewBroadcastHandler(
	broadcastRepo BroadcastRepository,
	conversationService *ConversationService,
	rdb *redis.Client,
	streamName string,
) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastRepo:       broadcastRepo,
		conversationService: conversationService,
		rdb:                 rdb,
		streamName:          streamName,
	}
}

// CreateBroadcast handles POST /broadcasts
// Creates the broadcast record and fans out messages via the existing Redis stream.
func (h *BroadcastHandler) CreateBroadcast(w http.ResponseWriter, r *http.Request) {
	senderID := r.Header.Get("X-User-Id")
	if senderID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.RecipientIDs) == 0 {
		http.Error(w, `{"error":"At least one recipient is required"}`, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, `{"error":"Message content is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.RecipientIDs) > 500 {
		http.Error(w, `{"error":"Maximum 500 recipients per broadcast"}`, http.StatusBadRequest)
		return
	}

	// Default tag
	if req.Tag == "" {
		req.Tag = "📢 Announcement"
	}

	// Rate limit: 60 seconds between broadcasts
	ctx := context.Background()
	lastTime, err := h.broadcastRepo.GetLastBroadcastTime(ctx, senderID)
	if err != nil {
		http.Error(w, `{"error":"Failed to check rate limit"}`, http.StatusInternalServerError)
		return
	}
	if !lastTime.IsZero() && time.Since(lastTime) < 60*time.Second {
		remaining := 60 - int(time.Since(lastTime).Seconds())
		w.Header().Set("Retry-After", strconv.Itoa(remaining))
		http.Error(w, `{"error":"Rate limit: please wait before sending another broadcast"}`, http.StatusTooManyRequests)
		return
	}

	// Create the broadcast record
	broadcast := &Broadcast{
		SenderID:     senderID,
		Content:      req.Content,
		Tag:          req.Tag,
		RecipientIDs: req.RecipientIDs,
		TotalCount:   len(req.RecipientIDs),
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	broadcast, err = h.broadcastRepo.Create(ctx, broadcast)
	if err != nil {
		http.Error(w, `{"error":"Failed to create broadcast"}`, http.StatusInternalServerError)
		return
	}

	// Return immediately (HTTP 202) — fan-out happens in background
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(broadcast)

	// Fan-out in background goroutine
	go h.executeBroadcast(broadcast)
}

// executeBroadcast sends the broadcast message to each recipient by publishing
// to the existing Redis stream. The chat_service worker will persist each message,
// deliver it via WebSocket, and send APNs push notifications automatically.
func (h *BroadcastHandler) executeBroadcast(broadcast *Broadcast) {
	ctx := context.Background()
	h.broadcastRepo.UpdateStatus(ctx, broadcast.ID, "sending")

	successCount := 0
	failureCount := 0

	for _, recipientID := range broadcast.RecipientIDs {
		// Skip if recipient is the sender (shouldn't happen, but guard)
		if recipientID == broadcast.SenderID {
			continue
		}

		// Get or create conversation
		conv, err := h.conversationService.GetOrCreateConversation(
			ctx, []string{broadcast.SenderID, recipientID}, "")
		if err != nil {
			log.Printf("[Broadcast %s] Failed to get/create conversation for %s: %v\n",
				broadcast.ID, recipientID, err)
			failureCount++
			continue
		}

		h.broadcastRepo.AddConversationID(ctx, broadcast.ID, conv.ID)

		// Build message and publish to Redis stream (reuses entire existing pipeline)
		msgReq := MessageRequest{
			ConversationID: conv.ID,
			SenderID:       broadcast.SenderID,
			Content:        broadcast.Content,
			Timestamp:      time.Now().Unix(),
			BroadcastID:    broadcast.ID,
			Tag:            broadcast.Tag,
		}
		data, err := json.Marshal(msgReq)
		if err != nil {
			log.Printf("[Broadcast %s] Failed to marshal message for %s: %v\n",
				broadcast.ID, recipientID, err)
			failureCount++
			continue
		}

		err = h.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: h.streamName,
			Values: map[string]interface{}{"data": string(data)},
		}).Err()
		if err != nil {
			log.Printf("[Broadcast %s] Failed to publish to stream for %s: %v\n",
				broadcast.ID, recipientID, err)
			failureCount++
			continue
		}

		successCount++

		// Throttle: 50ms between messages to prevent Redis/APNs spike
		time.Sleep(50 * time.Millisecond)
	}

	// Mark broadcast as completed with final counts
	if err := h.broadcastRepo.SetCompleted(ctx, broadcast.ID, successCount, failureCount); err != nil {
		log.Printf("[Broadcast %s] Failed to update completion status: %v\n", broadcast.ID, err)
	}

	log.Printf("[Broadcast %s] Completed: %d/%d success, %d failures\n",
		broadcast.ID, successCount, broadcast.TotalCount, failureCount)
}

// ListBroadcasts handles GET /broadcasts?limit=20&offset=0
func (h *BroadcastHandler) ListBroadcasts(w http.ResponseWriter, r *http.Request) {
	senderID := r.Header.Get("X-User-Id")
	if senderID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	ctx := context.Background()
	broadcasts, err := h.broadcastRepo.ListBySender(ctx, senderID, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"Failed to list broadcasts"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(broadcasts)
}

// GetBroadcast handles GET /broadcasts/{id}
func (h *BroadcastHandler) GetBroadcast(w http.ResponseWriter, r *http.Request) {
	senderID := r.Header.Get("X-User-Id")
	if senderID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	broadcastID := r.PathValue("id")
	if broadcastID == "" {
		http.Error(w, `{"error":"broadcast id is required"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	broadcast, err := h.broadcastRepo.FindByID(ctx, broadcastID)
	if err != nil {
		http.Error(w, `{"error":"Broadcast not found"}`, http.StatusNotFound)
		return
	}

	// Authorization: only the sender can view their broadcasts
	if broadcast.SenderID != senderID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(broadcast)
}
