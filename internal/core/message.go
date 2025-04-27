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

func NewMessage(body []byte, producerID string) *Message {
	return &Message{
		Body:        body,
		ProducerID:  producerID,
		DeliveredTo: make(map[string]bool),
		AckedBy:     make(map[string]bool),
		Metadata:    make(map[string]string),
	}
}
