package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/google/uuid"
)

type InMemoryRepo struct {
	topics map[string]*topicEntry
	mu     sync.RWMutex
}

type topicEntry struct {
	messages []*core.Message
	offsets  map[string]int // consumerID -> offset
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		topics: make(map[string]*topicEntry),
	}
}

func (m *InMemoryRepo) CreateTopic(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.topics[name]; exists {
		return fmt.Errorf("topic %q already exists", name)
	}

	m.topics[name] = &topicEntry{
		messages: []*core.Message{},
		offsets:  map[string]int{},
	}

	return nil
}

func (m *InMemoryRepo) Publish(topic string, msg *core.Message) error {
	m.mu.RLock()
	topicEntry, exists := m.topics[topic]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %q does not exist", topic)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	msg.Timestamp = time.Now()

	topicEntry.messages = append(topicEntry.messages, msg)

	return nil
}

func (m *InMemoryRepo) Fetch(topic, consumerID string, limit int) ([]*core.Message, error) {

}

func (m *InMemoryRepo) CommitOffset(topic, consumerID string, offset int) error {

}

func (m *InMemoryRepo) GetOffset(topic, consumerID string) (int, error) {

}
