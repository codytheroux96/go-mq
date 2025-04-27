package core

type Consumer struct {
	ID      string
	Inbox   chan *Message
	Offsets map[string]int // Topic name -> index of last read message
}

func NewConsumer(id string) *Consumer {
	return &Consumer{
		ID:      id,
		Inbox:   make(chan *Message, 10),
		Offsets: make(map[string]int),
	}
}
