package broker

import (
	"testing"
	"time"

	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/codytheroux96/go-mq/internal/repository"
)

func TestManager(t *testing.T) {
	tests := []struct {
		name            string
		action          string 
		topic           string
		consumer        string
		preCreateTopic  bool
		preSubscribe    bool
		expectErr       bool
		expectDelivery  bool
	}{
		{
			name:           "Subscribe to existing topic",
			action:         "Subscribe",
			topic:          "test-topic",
			consumer:       "c1",
			preCreateTopic: true,
			expectErr:      false,
		},
		{
			name:           "Subscribe to missing topic",
			action:         "Subscribe",
			topic:          "missing-topic",
			consumer:       "c1",
			preCreateTopic: false,
			expectErr:      true,
		},
		{
			name:           "Publish to topic with one subscriber",
			action:         "Publish",
			topic:          "pub-topic",
			consumer:       "c2",
			preCreateTopic: true,
			preSubscribe:   true,
			expectErr:      false,
			expectDelivery: true,
		},
		{
			name:           "Publish to missing topic",
			action:         "Publish",
			topic:          "ghost-topic",
			preCreateTopic: false,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewInMemoryRepo()
			manager := NewManager(repo)

			if tt.preCreateTopic {
				if err := repo.CreateTopic(tt.topic); err != nil {
					t.Fatalf("failed to pre-create topic: %v", err)
				}
			}

			var inbox <-chan *core.Message

			if tt.preSubscribe {
				var err error
				inbox, err = manager.Subscribe(tt.topic, tt.consumer)
				if err != nil {
					t.Fatalf("failed to pre-subscribe: %v", err)
				}
			}

			switch tt.action {
			case "Subscribe":
				ch, err := manager.Subscribe(tt.topic, tt.consumer)
				if tt.expectErr && err == nil {
					t.Errorf("expected error, got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.expectErr && ch == nil {
					t.Errorf("expected a channel but got nil")
				}

			case "Publish":
				msg := core.NewMessage([]byte("hello"), "p1")
				err := manager.Publish(tt.topic, msg)

				if tt.expectErr && err == nil {
					t.Errorf("expected error, got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if tt.expectDelivery {
					select {
					case received := <-inbox:
						if string(received.Body) != "hello" {
							t.Errorf("expected message body 'hello', got '%s'", string(received.Body))
						}
					case <-time.After(500 * time.Millisecond):
						t.Errorf("expected message delivery, but nothing was received")
					}
				}
			default:
				t.Fatalf("unsupported action: %s", tt.action)
			}
		})
	}
}