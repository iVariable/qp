package processor

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"qp"
)

// Stdout - output message to stdout and acknowledge it
// Used mainly for debugging
type Stdout struct {
	logger        *log.Entry
}

// Process - process job
func (l *Stdout) Process(job qp.IJob) error {
	l.logger.WithField("job", job).Debug("Processing job")
	fmt.Printf("[Stdout processor] Received message: %#v\n", job.GetMessage())
	if ackError := job.AckMessage(); ackError != nil {
		l.logger.WithField("error", ackError).Debug("Error on message acknowledge")
		return ackError
	}
	l.logger.Debug("Message acknowledged")
	return nil
}

// Configure - configure processor
func (l *Stdout) Configure(configuration map[string]interface{}) error {
	l.logger = log.WithFields(log.Fields{
		"type":      "processor",
		"processor": "Stdout",
	})
	l.logger.Info("Configuration loaded")
	return nil
}
