package queue
import (
	"qp"
	"fmt"
	"utils"
	"github.com/ActiveState/tail"
	"errors"
)

type Tail struct {
	configuration tailConfiguration
	t *tail.Tail
	messages chan *tail.Line
}

type tailConfiguration struct {
	Path string
}

func (q *Tail) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &q.configuration)

	t, err := tail.TailFile(q.configuration.Path, tail.Config{Follow: true})
	if err != nil {
		panic("Failed to tail file: "+q.configuration.Path)
	}
	q.t = t
	q.messages = make(chan *tail.Line)

	go func(){
		for line := range q.t.Lines {
			q.messages <- line
		}
	}()

	return nil
}

func (q *Tail) GetName() string {
	return "Tail"
}

func (q *Tail) Consume() (*qp.Message, error) {
	fmt.Println("[Tail]: Consume message")

	line := <- q.messages

	return &qp.Message{Id: line.Time.String(), Body: line.Text}, nil
}

func (q *Tail) Ack(message *qp.Message) error {
	return nil
}

func (q *Tail) Reject(message *qp.Message) error {
	return errors.New("Tail queue does NOT support Reject() method")
}

func (q *Tail) GetNumberOfMessages() (int, error) {
	return 9999999, nil
}