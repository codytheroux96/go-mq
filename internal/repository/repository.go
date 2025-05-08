package repository

import (
	"github.com/codytheroux96/go-mq/internal/core"
)

type Repository interface {
	CreateTopic(name string) error
	ListTopics() ([]string, error)
	DeleteTopic(name string) error
	Fetch(topic, consumerID string, limit int) ([]*core.Message, error)
	CommitOffset(topic, consumerID string, offset int) error
	GetOffset(topic, consumerID string) (int, error)
	Publish(topic string, msg *core.Message) error
}
