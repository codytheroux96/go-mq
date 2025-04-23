package repository

import (
	"sync"

	"github.com/codytheroux96/go-mq/internal/core"
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
	
}

func (m *InMemoryRepo) Publish(topic string, msg *core.Message) error {

}

func (m *InMemoryRepo) Fetch(topic, consumerID string, limit int) ([]*core.Message, error) {

}

func (m *InMemoryRepo) CommitOffset(topic, consumerID string, offset int) error {

}

func (m *InMemoryRepo) GetOffset(topic, consumerID string) (int, error) {

}