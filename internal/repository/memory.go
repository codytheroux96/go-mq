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
	messages    []*core.Message
	offsets     map[string]int // consumerID -> offset
	subscribers map[string]*core.Consumer
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

	for _, consumer := range topicEntry.subscribers {
		select {
		case consumer.Inbox <- msg:
			// sent successfully
		default:
			// inbox is full and we will skip
		}
	}

	return nil
}

func (m *InMemoryRepo) Fetch(topic, consumerID string, limit int) ([]*core.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topicEntry, exists := m.topics[topic]
	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topic)
	}

	offset := topicEntry.offsets[consumerID]

	end := offset + limit
	if end > len(topicEntry.messages) {
		end = len(topicEntry.messages)
	}

	if offset >= len(topicEntry.messages) {
		return []*core.Message{}, nil
	}

	return topicEntry.messages[offset:end], nil
}

func (m *InMemoryRepo) CommitOffset(topic, consumerID string, offset int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topicEntry, exists := m.topics[topic]
	if !exists {
		return fmt.Errorf("topic %q does not exist", topic)
	}

	if offset > len(topicEntry.messages) {
		return fmt.Errorf("cannot commit offset %d beyond the topic length %d", offset, len(topicEntry.messages))
	}

	topicEntry.offsets[consumerID] = offset
	return nil
}

func (m *InMemoryRepo) GetOffset(topic, consumerID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topicEntry, exists := m.topics[topic]
	if !exists {
		return 0, fmt.Errorf("topic %q does not exist", topic)
	}

	offset, ok := topicEntry.offsets[consumerID]
	if !ok {
		return 0, nil
	}

	return offset, nil
}

func (m *InMemoryRepo) Subscribe(topicName, consumerID string) (<-chan *core.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	topicEntry, exists := m.topics[topicName]
	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topicName)
	}

	if _, ok := topicEntry.subscribers[consumerID]; ok {
		return topicEntry.subscribers[consumerID].Inbox, nil
	}

	consumer := core.NewConsumer(consumerID)

	topicEntry.subscribers[consumerID] = consumer

	topicEntry.offsets[consumerID] = 0

	return nil, nil
}
