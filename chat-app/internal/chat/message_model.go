package chat

import "sync"

// Using a sync.Pool for messages to reduce garbage collector pressure.
var messagePool = sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

// Message defines the structure for websocket communication.
type Message struct {
	Type      string      `json:"type"`      // e.g., "private_message", "group_message", "user_list"
	Content   interface{} `json:"content"`   // Can be string for messages or []string for user list
	Sender    string      `json:"sender"`    // Username of the sender
	Recipient string      `json:"recipient"` // Username or Group name
}

func GetMessageFromPool() *Message {
	return messagePool.Get().(*Message)
}

func PutMessageInPool(m *Message) {
	m.Type = ""
	m.Content = ""
	m.Sender = ""
	m.Recipient = ""
	messagePool.Put(m)
}
