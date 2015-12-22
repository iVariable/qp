package processor

import (
	"fmt"
	"qp"
	log "github.com/Sirupsen/logrus"
)

type Stdout struct {
	configuration qp.Configuration
	logger *log.Entry
}

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

func (l *Stdout) Configure(configuration map[string]interface{}) error {
	l.logger = log.WithFields(log.Fields{
		"type": "processor",
		"processor": "Stdout",
	})
	l.logger.Info("Configuration loaded")
	return nil
}
