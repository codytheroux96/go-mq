package core

type Consumer struct {
	ID      string
	Inbox   chan *Message
	Offsets map[string]int // Topic name -> index of last read message
}
