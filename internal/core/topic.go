package core

import "sync"

type Topic struct {
	Name      string
	Messages  []*Message
	Consumers map[string]*Consumer
	Mu        sync.RWMutex
}
