package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/codytheroux96/go-mq/internal/app"
)

func setupTestServer() http.Handler {
	a := app.NewApplication()
	return Routes(a)
}

func makeRequest(ts http.Handler, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	ts.ServeHTTP(rr, req)
	return rr
}

func TestHandlers(t *testing.T) {
	ts := setupTestServer()

	t.Run("POST /topics", func(t *testing.T) {
		tests := []struct {
			name       string
			payload    map[string]string
			headers    map[string]string
			statusCode int
		}{
			{"Valid topic creation", map[string]string{"name": "t1"}, map[string]string{"Content-Type": "application/json"}, http.StatusCreated},
			{"Missing Content-Type", map[string]string{"name": "t2"}, nil, http.StatusUnsupportedMediaType},
			{"Empty topic name", map[string]string{"name": ""}, map[string]string{"Content-Type": "application/json"}, http.StatusBadRequest},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				body, _ := json.Marshal(tc.payload)
				rr := makeRequest(ts, http.MethodPost, "/topics", bytes.NewReader(body), tc.headers)
				if rr.Code != tc.statusCode {
					t.Errorf("expected %d, got %d", tc.statusCode, rr.Code)
				}
			})
		}
	})

	t.Run("GET /topics", func(t *testing.T) {
		rr := makeRequest(ts, http.MethodGet, "/topics", nil, nil)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /topics/{topic}", func(t *testing.T) {
		rr := makeRequest(ts, http.MethodDelete, "/topics/t1", nil, nil)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		rr = makeRequest(ts, http.MethodDelete, "/topics/unknown", nil, nil)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("POST /publish/{topic}", func(t *testing.T) {
		_ = makeRequest(ts, http.MethodPost, "/topics", bytes.NewReader([]byte(`{"name":"publish-topic"}`)), map[string]string{"Content-Type": "application/json"})
		body := `{"body": "message content", "producer_id": "p1"}`
		rr := makeRequest(ts, http.MethodPost, "/publish/publish-topic", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
		if rr.Code != http.StatusAccepted {
			t.Errorf("expected 202, got %d", rr.Code)
		}
	})

	t.Run("POST /subscribe", func(t *testing.T) {
		_ = makeRequest(ts, http.MethodPost, "/topics", bytes.NewReader([]byte(`{"name":"sub-topic"}`)), map[string]string{"Content-Type": "application/json"})
		body := `{"topic": "sub-topic", "consumer_id": "c1"}`
		rr := makeRequest(ts, http.MethodPost, "/subscribe", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("GET /subscribe/{topic}", func(t *testing.T) {
		_ = makeRequest(ts, http.MethodPost, "/subscribe", strings.NewReader(`{"topic": "sub-topic", "consumer_id": "c2"}`), map[string]string{"Content-Type": "application/json"})
		rr := makeRequest(ts, http.MethodGet, "/subscribe/sub-topic?consumer_id=c2", nil, nil)
		if rr.Code != http.StatusNoContent && rr.Code != http.StatusOK {
			t.Errorf("expected 204 or 200, got %d", rr.Code)
		}
	})

	t.Run("POST /ack", func(t *testing.T) {
		rr := makeRequest(ts, http.MethodPost, "/ack", strings.NewReader(`{"topic":"sub-topic","consumer_id":"c1","message_id":"nonexistent"}`), map[string]string{"Content-Type": "application/json"})
		if rr.Code != http.StatusNotFound && rr.Code != http.StatusForbidden {
			t.Errorf("expected 404 or 403, got %d", rr.Code)
		}
	})

	t.Run("GET /health", func(t *testing.T) {
		rr := makeRequest(ts, http.MethodGet, "/health", nil, nil)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("GET /fetch", func(t *testing.T) {
		rr := makeRequest(ts, http.MethodGet, "/fetch", nil, map[string]string{
			"X-Topic":       "sub-topic",
			"X-Consumer-ID": "c1",
			"X-Limit":       strconv.Itoa(5),
			"X-Commit":      "true",
		})
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
