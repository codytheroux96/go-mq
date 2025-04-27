package core

import "time"

type Producer struct {
	ID         string
	Metadata   map[string]string
	LastActive time.Time
	MessageIDs []string
}

func NewProducer(id string) *Producer {
	return &Producer{
		ID:         id,
		Metadata:   make(map[string]string),
		LastActive: time.Now(),
		MessageIDs: []string{},
	}
}
