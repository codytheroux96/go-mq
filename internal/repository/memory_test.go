package repository

import (
	"fmt"
	"testing"

	"github.com/codytheroux96/go-mq/internal/core"
)

func TestInMemoryRepo(t *testing.T) {
	repo := NewInMemoryRepo()

	tests := []struct {
		name      string
		action    func() (any, error)
		expectErr bool
		expectVal any
	}{
		{
			name: "Create new topic",
			action: func() (any, error) {
				return nil, repo.CreateTopic("test-topic")
			},
			expectErr: false,
		},
		{
			name: "Create duplicate topic",
			action: func() (any, error) {
				return nil, repo.CreateTopic("test-topic")
			},
			expectErr: true,
		},
		{
			name: "Publish message to topic",
			action: func() (any, error) {
				msg := &core.Message{Body: []byte("Test message 1")}
				return nil, repo.Publish("test-topic", msg)
			},
			expectErr: false,
		},
		{
			name: "Publish another message",
			action: func() (any, error) {
				msg := &core.Message{Body: []byte("Test message 2")}
				return nil, repo.Publish("test-topic", msg)
			},
			expectErr: false,
		},
		{
			name: "Fetch first message batch",
			action: func() (any, error) {
				return repo.Fetch("test-topic", "consumer-1", 1)
			},
			expectErr: false,
			expectVal: 1, // expecting 1 message fetched
		},
		{
			name: "Get initial offset",
			action: func() (any, error) {
				return repo.GetOffset("test-topic", "consumer-1")
			},
			expectErr: false,
			expectVal: 0, // still at offset 0 until commit
		},
		{
			name: "Commit offset after processing",
			action: func() (any, error) {
				return nil, repo.CommitOffset("test-topic", "consumer-1", 1)
			},
			expectErr: false,
		},
		{
			name: "Get updated offset",
			action: func() (any, error) {
				return repo.GetOffset("test-topic", "consumer-1")
			},
			expectErr: false,
			expectVal: 1, // now offset should be 1
		},
		{
			name: "Fetch second message batch",
			action: func() (any, error) {
				return repo.Fetch("test-topic", "consumer-1", 1)
			},
			expectErr: false,
			expectVal: 1, // should fetch the next 1 message
		},
		{
			name: "Subscribe to topic",
			action: func() (any, error) {
				return repo.Subscribe("test-topic", "consumer-2")
			},
			expectErr: false,
		},
		{
			name: "Publish delivers to live subscriber",
			action: func() (any, error) {
				ch, err := repo.Subscribe("test-topic", "consumer-3")
				if err != nil {
					return nil, err
				}

				msg := core.NewMessage([]byte("live message"), "producer-1")
				err = repo.Publish("test-topic", msg)
				if err != nil {
					return nil, err
				}

				select {
				case received := <-ch:
					if string(received.Body) != "live message" {
						return nil, fmt.Errorf("expected 'live message', got '%s'", string(received.Body))
					}
					return "delivered", nil
				default:
					return nil, fmt.Errorf("no message received on subscriber channel")
				}
			},
			expectErr: false,
			expectVal: "delivered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.action()

			if tt.expectErr && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectVal != nil {
				switch expected := tt.expectVal.(type) {
				case int:
					switch v := got.(type) {
					case int:
						if v != expected {
							t.Fatalf("expected offset %d, got %d", expected, v)
						}
					case []*core.Message:
						if len(v) != expected {
							t.Fatalf("expected %d messages, got %d", expected, len(v))
						}
					default:
						t.Fatalf("unexpected return type: %T", got)
					}
				case string:
					if gotStr, ok := got.(string); !ok || gotStr != expected {
						t.Fatalf("expected value %q, got %v", expected, got)
					}
				default:
					t.Fatalf("unsupported expectVal type: %T", tt.expectVal)
				}
			}
		})
	}
}
