package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codytheroux96/go-mq/internal/app"
	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/codytheroux96/go-mq/internal/repository"
)

type Handler struct {
	App *app.Application
}

func (h *Handler) HandleTopics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleCreateTopic(w, r)
	case http.MethodGet:
		h.HandleListTopics(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleCreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		h.App.Logger.Warn("invalid content-type received", "received", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		h.App.Logger.Error("failed to decode request body or request body is missing name", "error", err)
		http.Error(w, "invalid payload in request", http.StatusBadRequest)
		return
	}

	if err := h.App.Repo.CreateTopic(req.Name); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.App.Logger.Warn("attempt to create duplicate topic was made", "topic", req.Name)
			http.Error(w, "cannot create topic - topic already exists", http.StatusConflict)
			return
		}
		h.App.Logger.Error("failed to create toic,", "topic", req.Name)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	h.App.Logger.Info("topic created", "topic", req.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "topic created successfully"})
}

func (h *Handler) HandleListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.App.Repo.ListTopics()
	if err != nil {
		h.App.Logger.Error("failed to fetch list topics", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.App.Logger.Info("listing all topics", "count", len(topics))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]string{"topics": topics})
}

func (h *Handler) HandleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.App.Logger.Warn("http method not allowed for deleting topic", "method", r.Method)
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := strings.TrimPrefix(r.URL.Path, "/topics/")
	if topicName == "" {
		h.App.Logger.Warn("missing topic name in delete request")
		http.Error(w, "topic name is required for delete request", http.StatusBadRequest)
		return
	}

	if err := h.App.Repo.DeleteTopic(topicName); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.App.Logger.Error("attempt to delete topic that does not exist", "topic", topicName)
			http.Error(w, "topic requested to be deleted does not exist", http.StatusNotFound)
			return
		}
		h.App.Logger.Error("failed to delete topic", "topic", topicName)
		http.Error(w, "failed to delete topic", http.StatusInternalServerError)
		return
	}

	h.App.Logger.Info("topic was successfully deleted", "topic", topicName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "topic deleted successfully"})
}

func (h *Handler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.App.Logger.Warn("http method not allowed for publishing a message to a topic", "method", r.Method)
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := strings.TrimPrefix(r.URL.Path, "/publish/")
	if topicName == "" {
		h.App.Logger.Warn("missing topic name in request to publish to a topic")
		http.Error(w, "topic name is required to publish to a topic", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		h.App.Logger.Warn("invalid content-type received", "received", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req struct {
		Body       string `json:"body"`
		ProducerID string `json:"producer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" || req.ProducerID == "" {
		h.App.Logger.Error("failed to decode request. check that request body and producer_id are both in request", "error", err)
		http.Error(w, "invalid payload in request", http.StatusBadRequest)
		return
	}

	msg := core.NewMessage([]byte(req.Body), req.ProducerID)
	if err := h.App.Repo.Publish(topicName, msg); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.App.Logger.Error("attempting to publish to a topic that does not exist", "topic", topicName)
			http.Error(w, "topic requested to publish to does not exist", http.StatusNotFound)
			return
		}
		h.App.Logger.Error("failed to publish to topic", "topic", topicName)
		http.Error(w, "failed to publish to topic", http.StatusInternalServerError)
		return
	}

	h.App.Logger.Info("message publish successfully to topic", "topic", topicName, "producer_id", req.ProducerID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "message published successfully"})
}

func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.App.Logger.Warn("http method not allowed for subscribing to a topic", "method", r.Method)
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := strings.TrimPrefix(r.URL.Path, "/subscribe/")
	if topicName == "" {
		h.App.Logger.Warn("missing topic name in subscribe request")
		http.Error(w, "topic name is required in order to subscribe", http.StatusBadRequest)
		return
	}

	consumerID := r.URL.Query().Get("consumer_id")
	if consumerID == "" {
		h.App.Logger.Warn("missing consumer ID in subscribe request")
		http.Error(w, "cosumer_id is required in order to subscribe", http.StatusBadRequest)
		return
	}

	inbox, err := h.App.Repo.Subscribe(topicName, consumerID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.App.Logger.Warn("subscribe was attempted on a non-existent topic", "topic", topicName)
			http.Error(w, "topic does not exist", http.StatusNotFound)
			return
		}
		h.App.Logger.Error("failed to subscribe to topic", "topic", topicName, "error", err)
		http.Error(w, "failed to subscribe to topic", http.StatusInternalServerError)
		return
	}

	select {
	case msg := <-inbox:
		h.App.Logger.Info("delivered message to consumer", "topic", topicName, "consumer", consumerID, "message_id", msg.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"body":        string(msg.Body),
			"producer_id": msg.ProducerID,
			"timestamp":   msg.Timestamp,
			"message_id":  msg.ID,
		})
	case <-time.After(10 * time.Second):
		h.App.Logger.Info("subscribe timeout: no messages", "topic", topicName, "consumer", consumerID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) HandleRegisterConsumer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.App.Logger.Warn("http method not allowed for registering a consumer", "method", r.Method)
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		h.App.Logger.Warn("invalid content-type received", "received", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req struct {
		Topic      string `json:"topic"`
		ConsumerID string `json:"consumer_id"`
		Offset     *int   `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" || req.ConsumerID == "" {
		h.App.Logger.Error("invalid consumer registration request", "error", err)
		http.Error(w, "invalid payload in request", http.StatusBadRequest)
		return
	}

	_, err := h.App.Repo.Subscribe(req.Topic, req.ConsumerID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			h.App.Logger.Warn("subscribe attempted on non-existent topic", "topic", req.Topic)
			http.Error(w, "topic does not exist", http.StatusNotFound)
			return
		}
		h.App.Logger.Error("failed to subscribe consumer", "topic", req.Topic, "consumer", req.ConsumerID, "error", err)
		http.Error(w, "failed to register consumer", http.StatusInternalServerError)
		return
	}

	if req.Offset != nil {
		if err := h.App.Repo.CommitOffset(req.Topic, req.ConsumerID, *req.Offset); err != nil {
			h.App.Logger.Warn("failed to set custom offset after consumer registration", "topic", req.Topic, "consumer", req.ConsumerID, "offset", *req.Offset, "error", err)
			http.Error(w, "consumer registered but failed to set offset", http.StatusBadRequest)
			return
		}
		h.App.Logger.Info("custom offset successfully sete during consumer registration", "topic", req.Topic, "consumer", req.ConsumerID, "offset", *req.Offset)
	}

	h.App.Logger.Info("consumer subscribed successfully", "topic", req.Topic, "consumer", req.ConsumerID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "consumer subscribed successfully",
	})
}

func (h *Handler) HandleAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.App.Logger.Warn("http method is not allowed for acknowledging", "method", r.Method)
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		h.App.Logger.Warn("invalid content-type for acknowledging", "received", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req struct {
		Topic      string `json:"topic"`
		ConsumerID string `json:"consumer_id"`
		MessageID  string `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" || req.ConsumerID == "" || req.MessageID == "" {
		h.App.Logger.Error("invalid ack request payload", "error", err)
		http.Error(w, "invalid payload in request", http.StatusBadRequest)
		return
	}

	repo := h.App.Repo

	memoryRepo, ok := repo.(*repository.InMemoryRepo)
	if !ok {
		h.App.Logger.Error("ack only supported in an in-memory repo")
		http.Error(w, "ack not supported in current backend", http.StatusInternalServerError)
		return
	}

	memoryRepo.Mu.Lock()
	defer memoryRepo.Mu.Unlock()

	topicEntry, exists := memoryRepo.Topics[req.Topic]
	if !exists {
		h.App.Logger.Warn("ack failed: topic not found", "topic", req.Topic)
		http.Error(w, "topic not found", http.StatusNotFound)
		return
	}

	var targetMsg *core.Message
	for _, msg := range topicEntry.Messages {
		if msg.ID == req.MessageID {
			targetMsg = msg
			break
		}
	}

	if targetMsg == nil {
		h.App.Logger.Warn("ack failed: message not found", "message_id", req.MessageID)
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}

	if !targetMsg.DeliveredTo[req.ConsumerID] {
		h.App.Logger.Warn("ack rejected: message not delivered to consumer", "message_id", req.MessageID, "consumer", req.ConsumerID)
		http.Error(w, "message not delivered to this consumer", http.StatusForbidden)
		return
	}

	if targetMsg.AckedBy[req.ConsumerID] {
		h.App.Logger.Info("duplicate ack received", "message_id", req.MessageID, "consumer", req.ConsumerID)
	} else {
		targetMsg.AckedBy[req.ConsumerID] = true
		h.App.Logger.Info("message acknowledged", "message_id", req.MessageID, "consumer", req.ConsumerID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "message acknowledged successfully",
	})
}

func (h *Handler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "go-mq",
	})
}

func (h *Handler) HandleFetchMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.App.Logger.Warn("http method not allowed for fetch", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topic := r.Header.Get("X-Topic")
	consumerID := r.Header.Get("X-Consumer-ID")
	limitStr := r.Header.Get("X-Limit")
	offsetStr := r.Header.Get("X-Offset")
	commit := strings.ToLower(r.Header.Get("X-Commit")) == "true"

	if topic == "" || consumerID == "" {
		h.App.Logger.Warn("missing required fetch headers:", "topic", topic, "consumerID", consumerID)
		http.Error(w, "X-Topic and X-Consumer-ID headers are required", http.StatusBadRequest)
		return
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := -1
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var startOffset int
	var err error
	if offset >= 0 {
		startOffset = offset
	} else {
		startOffset, err = h.App.Repo.GetOffset(topic, consumerID)
		if err != nil {
			h.App.Logger.Error("failed to get offset", "topic", topic, "consumer", consumerID, "error", err)
			http.Error(w, "failed to retrieve offset", http.StatusInternalServerError)
			return
		}
	}

	messages, err := h.App.Repo.Fetch(topic, consumerID, limit)
	if err != nil {
		h.App.Logger.Error("failed to fetch messages", "topic", topic, "consumer", consumerID, "error", err)
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
	}

	if commit {
		if err := h.App.Repo.CommitOffset(topic, consumerID, startOffset+len(messages)); err != nil {
			h.App.Logger.Warn("failed to auto-commit offset", "topic", topic, "consumer", consumerID, "error", err)
		} else {
			h.App.Logger.Info("fetched and committed messages", "topic", topic, "consumer", consumerID, "new_offset", startOffset+len(messages))
		}
	} else {
		h.App.Logger.Info("fetched messages without committing", "topic", topic, "consumer", consumerID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
