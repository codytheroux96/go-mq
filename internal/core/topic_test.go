package core

import (
	"testing"
)

func TestTopic(t *testing.T) {
	topic := NewTopic("test-topic")

	consumer1 := NewConsumer("consumer-1")
	consumer2 := NewConsumer("consumer-2")

	tests := []struct {
		name      string
		action    func() error
		expectErr bool
	}{
		{
			name: "Add first consumer successfully",
			action: func() error {
				return topic.AddConsumer(consumer1)
			},
			expectErr: false,
		},
		{
			name: "Add duplicate consumer should fail",
			action: func() error {
				return topic.AddConsumer(consumer1)
			},
			expectErr: true,
		},
		{
			name: "Add second consumer successfully",
			action: func() error {
				return topic.AddConsumer(consumer2)
			},
			expectErr: false,
		},
		{
			name: "Remove first consumer successfully",
			action: func() error {
				topic.RemoveConsumer(consumer1.ID)
				return nil
			},
			expectErr: false,
		},
		{
			name: "Broadcast message to active consumers",
			action: func() error {
				msg := NewMessage([]byte("hello world"), "producer-1")
				topic.Broadcast(msg)

				// Only consumer2 should have received the message (since consumer1 was removed)
				select {
				case received := <-consumer2.Inbox:
					if string(received.Body) != "hello world" {
						t.Errorf("Expected message body 'hello world', got '%s'", string(received.Body))
					}
				default:
					t.Errorf("Expected consumer2 to receive message but inbox was empty")
				}

				return nil
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
