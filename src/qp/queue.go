package qp

type ConsumableQueue interface {
	GetName() string
	Configure(config interface{}) error
	Consume() (*Message, error)
	Ack(message *Message) error
	Reject(message *Message) error
	GetNumberOfMessages() (int, error)
}

type Message struct {
	Id interface{}
	Body interface{}
}