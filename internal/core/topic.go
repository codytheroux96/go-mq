package core

import "sync"

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
