package broker

import (
	"fmt"
	"sync"
	"time"

	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/codytheroux96/go-mq/internal/repository"
	"github.com/google/uuid"
)

type Manager struct {
	Repo   repository.Repository
	Topics map[string]*core.Topic
	Mu     sync.RWMutex
}

func NewManager(repo repository.Repository) *Manager {
	return &Manager{
		Repo:   repo,
		Topics: make(map[string]*core.Topic),
	}
}

func (b *Manager) Subscribe(topicName, consumerID string) (<-chan *core.Message, error) {
	b.Mu.Lock()
	defer b.Mu.Unlock()

	if _, ok := b.Topics[topicName]; !ok {
		topics, err := b.Repo.ListTopics()
		if err != nil {
			return nil, fmt.Errorf("failed to verify topic existence: %w", err)
		}
		found := false
		for _, t := range topics {
			if t == topicName {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("topic %q does not exist", topicName)
		}
		b.Topics[topicName] = core.NewTopic(topicName)
	}

	topic := b.Topics[topicName]

	consumer := core.NewConsumer(consumerID)
	topic.Consumers[consumerID] = consumer

	offset, err := b.Repo.GetOffset(topicName, consumerID)
	if err != nil {
		_ = b.Repo.CommitOffset(topicName, consumerID, 0)
	} else {
		_ = offset
	}

	return consumer.Inbox, nil
}

func (b *Manager) Publish(topic string, msg *core.Message) error {
	b.Mu.Lock()
	defer b.Mu.Unlock()

	if _, ok := b.Topics[topic]; !ok {
		topics, err := b.Repo.ListTopics()
		if err != nil {
			return fmt.Errorf("failed to verify topic existence: %w", err)
		}
		found := false
		for _, t := range topics {
			if t == topic {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("topic %q does not exist", topic)
		}
		b.Topics[topic] = core.NewTopic(topic)
	}

	topicEntry := b.Topics[topic]

	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	msg.Timestamp = time.Now()

	if err := b.Repo.Publish(topic, msg); err != nil {
		return err
	}

	for consumerID, consumer := range topicEntry.Consumers {
		select {
		case consumer.Inbox <- msg:
			msg.DeliveredTo[consumerID] = true
		default:
			// inbox is full — skip delivery
		}
	}

	return nil
}
