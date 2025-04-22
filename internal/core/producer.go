package core

import "time"

type Producer struct {
	ID         string
	Metadata   map[string]string
	LastActive time.Time
	MessageIDs []string
}
