package queue

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/iVariable/qp/src/qp"
	"github.com/iVariable/qp/src/utils"
	"math/rand"
	"time"
)

// Dummy - dummy queue which generates static messages with randomizable delay
// Used for debugging
type Dummy struct {
	configuration dummyConfiguration
	logger        *log.Entry
}

type dummyConfiguration struct {
	RandomSleepDelay int
}

// GetName returns queue name
func (q *Dummy) GetName() string {
	return "Dummy queue"
}

// Consume consumes a message
func (q *Dummy) Consume() (qp.IMessage, error) {
	q.logger.Debug("Message consume")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(q.configuration.RandomSleepDelay)) * time.Millisecond) //TODO make configurable

	return &qp.Message{
		ID:   time.Now(),
		Body: fmt.Sprintf("Dummy message generated | break at %s", time.Now())}, nil
}

// Ack acknowledges a message
func (q *Dummy) Ack(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message acknowledged")
	return nil
}

// Configure configure queue
func (q *Dummy) Configure(configuration map[string]interface{}) error {
	q.logger = log.WithFields(log.Fields{
		"type":  "queue",
		"queue": "Dummy",
	})
	q.logger.Debug("Reading configuration")
	utils.FillStruct(configuration, &q.configuration)
	if q.configuration.RandomSleepDelay < 0 {
		return errors.New("RandomSleepDelay should be >= 0")
	}
	q.logger.WithField("configuration", q.configuration).Info("Configuration loaded")
	return nil
}

// Reject reject a message
func (q *Dummy) Reject(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message rejected")
	return nil
}

// GetNumberOfMessages returns number of messages
func (q *Dummy) GetNumberOfMessages() (int, error) {
	return 9999, nil
}
