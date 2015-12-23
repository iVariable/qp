package queue

import (
	"errors"
	"qp"
	"sync"
	"utils"

	"github.com/ActiveState/tail"
	log "github.com/Sirupsen/logrus"
)

type Tail struct {
	configuration tailConfiguration
	t             *tail.Tail
	messages      chan *tail.Line
	once          sync.Once
	startTailing  func()
	logger		  *log.Entry
}

type tailConfiguration struct {
	Path string
}

func (q *Tail) Configure(configuration map[string]interface{}) error {
	q.logger = log.WithFields(log.Fields{
		"type": "queue",
		"queue": "Tail",
	})
	q.logger.Debug("Reading configuration")
	utils.FillStruct(configuration, &q.configuration)

	q.messages = make(chan *tail.Line)

	q.startTailing = func() {
		t, err := tail.TailFile(q.configuration.Path, tail.Config{Follow: true})
		if err != nil {
			q.logger.WithField("file", q.configuration.Path).Fatal("Failed to tail file")
			utils.Quitf(utils.EXITCODE_RUNTIME_ERROR, "Failed to tail file %s", q.configuration.Path)
		}
		q.t = t

		go func() {
			for line := range q.t.Lines {
				q.messages <- line
			}
		}()
	}

	q.logger.WithField("configuration", q.configuration).Info("Configuration loaded")
	return nil
}

func (q *Tail) GetName() string {
	return "Tail"
}

func (q *Tail) Consume() (qp.IMessage, error) {
	q.once.Do(q.startTailing)
	q.logger.Debug("Message consume")

	line := <-q.messages

	return &qp.Message{Id: line.Time.String(), Body: line.Text}, nil
}

func (q *Tail) Ack(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message acknowledged")
	return nil
}

func (q *Tail) Reject(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message rejected")
	return errors.New("Tail queue does NOT support Reject() method")
}

func (q *Tail) GetNumberOfMessages() (int, error) {
	return 9999999, nil
}
