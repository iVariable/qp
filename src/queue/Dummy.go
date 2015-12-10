package queue
import (
	"qp"
	"time"
	"fmt"
	"math/rand"
	"utils"
	"errors"
)

type Dummy struct {
	configuration dummyConfiguration
}

type dummyConfiguration struct {
	RandomSleepDelay int
}

func (q *Dummy) GetName() string {
	return "Dummy queue"
}

func (q *Dummy) Consume() (qp.IMessage, error) {
	fmt.Println("[Dummy Queue]: Consume message")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(q.configuration.RandomSleepDelay)) * time.Millisecond) //TODO make configurable

	return &qp.Message{
		Id: time.Now(),
		Body: fmt.Sprintf("Dummy message generated at %s", time.Now())}, nil
}

func (q *Dummy) Ack(message *qp.IMessage) error {
	fmt.Println("[Dummy Queue]: Ack message")
	return nil
}

func (q *Dummy) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &q.configuration)
	if q.configuration.RandomSleepDelay < 0 {
		return errors.New("RandomSleepDelay should be >= 0")
	}
	return nil
}

func (q *Dummy) Reject(message *qp.IMessage) error {
	fmt.Println("[Dummy Queue]: Reject message")
	return nil
}

func (q *Dummy) GetNumberOfMessages() (int, error) {
	return 9999, nil
}