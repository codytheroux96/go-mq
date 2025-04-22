package core

import "time"

type Message struct {
	ID          string
	Body        []byte
	Timestamp   time.Time
	ProducerID  string
	DeliveredTo map[string]bool
	AckedBy     map[string]bool
	Metadata    map[string]string
}
