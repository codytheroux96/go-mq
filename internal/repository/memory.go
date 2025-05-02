package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/codytheroux96/go-mq/internal/core"

	"github.com/google/uuid"
)

type InMemoryRepo struct {
	Topics map[string]*topicEntry
	Mu     sync.RWMutex
}

type topicEntry struct {
	Messages    []*core.Message
	Offsets     map[string]int // consumerID -> offset
	Subscribers map[string]*core.Consumer
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		Topics: make(map[string]*topicEntry),
	}
}

func (m *InMemoryRepo) CreateTopic(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Topics[name]; exists {
		return fmt.Errorf("topic %q already exists", name)
	}

	m.Topics[name] = &topicEntry{
		Messages:    []*core.Message{},
		Offsets:     map[string]int{},
		Subscribers: map[string]*core.Consumer{},
	}

	return nil
}

func (m *InMemoryRepo) ListTopics() ([]string, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	topics := make([]string, 0, len(m.Topics))
	for name := range m.Topics {
		topics = append(topics, name)
	}

	return topics, nil
}

func (m *InMemoryRepo) Publish(topic string, msg *core.Message) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	topicEntry, exists := m.Topics[topic]
	if !exists {
		return fmt.Errorf("topic %q does not exist", topic)
	}

	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	msg.Timestamp = time.Now()

	topicEntry.Messages = append(topicEntry.Messages, msg)

	for consumerID, consumer := range topicEntry.Subscribers {
		select {
		case consumer.Inbox <- msg:
			msg.DeliveredTo[consumerID] = true
		default:
			// inbox is full and we will skip
		}
	}

	return nil
}

func (m *InMemoryRepo) DeleteTopic(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.Topics[name]; !exists {
		return fmt.Errorf("topic %q does not exist", name)
	}

	delete(m.Topics, name)
	return nil
}

func (m *InMemoryRepo) Fetch(topic, consumerID string, limit int) ([]*core.Message, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	topicEntry, exists := m.Topics[topic]
	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topic)
	}

	offset := topicEntry.Offsets[consumerID]

	end := offset + limit
	if end > len(topicEntry.Messages) {
		end = len(topicEntry.Messages)
	}

	if offset >= len(topicEntry.Messages) {
		return []*core.Message{}, nil
	}

	return topicEntry.Messages[offset:end], nil
}

func (m *InMemoryRepo) CommitOffset(topic, consumerID string, offset int) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	topicEntry, exists := m.Topics[topic]
	if !exists {
		return fmt.Errorf("topic %q does not exist", topic)
	}

	if offset > len(topicEntry.Messages) {
		return fmt.Errorf("cannot commit offset %d beyond the topic length %d", offset, len(topicEntry.Messages))
	}

	topicEntry.Offsets[consumerID] = offset
	return nil
}

func (m *InMemoryRepo) GetOffset(topic, consumerID string) (int, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	topicEntry, exists := m.Topics[topic]
	if !exists {
		return 0, fmt.Errorf("topic %q does not exist", topic)
	}

	offset, ok := topicEntry.Offsets[consumerID]
	if !ok {
		return 0, nil
	}

	return offset, nil
}

func (m *InMemoryRepo) Subscribe(topicName, consumerID string) (<-chan *core.Message, error) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	topicEntry, exists := m.Topics[topicName]
	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topicName)
	}

	if _, ok := topicEntry.Subscribers[consumerID]; ok {
		return topicEntry.Subscribers[consumerID].Inbox, nil
	}

	consumer := core.NewConsumer(consumerID)

	topicEntry.Subscribers[consumerID] = consumer

	topicEntry.Offsets[consumerID] = 0

	return consumer.Inbox, nil
}
