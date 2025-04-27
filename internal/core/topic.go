package core

import (
	"fmt"
	"sync"
)

type Topic struct {
	Name      string
	Messages  []*Message
	Consumers map[string]*Consumer
	Mu        sync.RWMutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		Name:      name,
		Messages:  []*Message{},
		Consumers: make(map[string]*Consumer),
	}
}

func (t *Topic) AddConsumer(consumer *Consumer) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if _, exists := t.Consumers[consumer.ID]; exists {
		return fmt.Errorf("consumer %q is already subscribed", consumer.ID)
	}

	t.Consumers[consumer.ID] = consumer
	return nil
}

func (t *Topic) RemoveConsumer(consumerID string) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	delete(t.Consumers, consumerID)
}

func (t *Topic) Broadcast(msg *Message) {
	t.Mu.RLock()
	defer t.Mu.RUnlock()

	for _, consumer := range t.Consumers {
		select {
		case consumer.Inbox <- msg:
			// means it was sent successfully
		default:
			// else inbox is full and we will skip
		}
	}
}
