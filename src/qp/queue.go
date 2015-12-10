package qp

type ConsumableQueue interface {
	GetName() string
	Configure(configuration map[string]interface{}) error
	Consume() (*Message, error)
	Ack(message *Message) error
	Reject(message *Message) error
	GetNumberOfMessages() (int, error)
}

type Message struct {
	Id interface{}
	Body interface{}
}