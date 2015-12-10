package qp
import "encoding/json"

type ConsumableQueue interface {
	GetName() string
	Configure(configuration map[string]interface{}) error
	Consume() (IMessage, error)
	Ack(message *IMessage) error
	Reject(message *IMessage) error
	GetNumberOfMessages() (int, error)
}

type Message struct {
	Id interface{}
	Body interface{}
}

type IMessage interface {
	Serialize() (string, error)
	GetId() interface{}
	GetBody() interface{}
}

func (m *Message) GetId() interface{} {
	return m.Id
}

func (m *Message) GetBody() interface{} {
	return m.Body
}

func (m *Message) Serialize() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}