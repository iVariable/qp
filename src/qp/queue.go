package qp

import "encoding/json"

// IConsumableQueue Consumable queue interface
type IConsumableQueue interface {
	GetName() string
	Configure(configuration map[string]interface{}) error
	Consume() (IMessage, error)
	Ack(message IMessage) error
	Reject(message IMessage) error
	GetNumberOfMessages() (int, error)
}

// IMessage - message interface
type IMessage interface {
	Serialize() (string, error)
	GetID() interface{}
	GetBody() interface{}
	GetRaw() string
}

// Message simple message struct
type Message struct {
	ID   interface{}
	Body interface{}
	Raw  string
}

// GetID returns message id
func (m *Message) GetID() interface{} {
	return m.ID
}

// GetBody returns message body
func (m *Message) GetBody() interface{} {
	return m.Body
}

func (m *Message) GetRaw() string {
	return m.Raw
}

// Serialize returns serialized representation of message
func (m *Message) Serialize() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
