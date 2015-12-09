package queue
import (
	"qp"
	"time"
	"fmt"
	"math/rand"
)

type Dummy struct {
	configuration qp.Configuration
}

func (q *Dummy) GetName() string {
	return "Dummy queue"
}

func (q *Dummy) Consume() (*qp.Message, error) {
	fmt.Println("[Dummy Queue]: Consume message")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(1000)) * time.Millisecond) //TODO make configurable

	if r.Intn(10) == 1 {
		fmt.Println("[Dummy Queue]: No messages to consume, sending nil")
		return nil, nil
	}

	return &qp.Message{
		Id: time.Now(),
		Body: fmt.Sprintf("Dummy message generated at %s", time.Now())}, nil
}

func (q *Dummy) Ack(message *qp.Message) error {
	fmt.Println("[Dummy Queue]: Ack message")
	return nil
}

func (q *Dummy) Configure(config interface{}) error {
	return nil
}

func (q *Dummy) Reject(message *qp.Message) error {
	fmt.Println("[Dummy Queue]: Reject message")
	return nil
}

func (q *Dummy) GetNumberOfMessages() (int, error) {
	return 9999, nil
}