package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/codytheroux96/go-mq/internal/app"
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
