package processor

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/iVariable/qp/src/qp"
	"github.com/iVariable/qp/src/utils"
	"os/exec"
	"strings"
)

// Shell - run custom shell command for message processing
// Acknowledge message in case exit code = 0
// Any other code - reject
type Shell struct {
	configuration shellConfiguration
	logger        *log.Entry
}

type shellConfiguration struct {
	Command            string
	MessagePlaceholder string
	EchoOutput         bool
	SendRaw            bool
}

// Process - Process job
func (l *Shell) Process(job qp.IJob) error {
	l.logger.WithField("job", job).Debug("Processing job")
	var msg string
	var err error
	if l.configuration.SendRaw {
		msg = job.GetMessage().GetRaw()
	} else {
		msg, err = job.GetMessage().Serialize()
		if err != nil {
			l.logger.WithField("error", err).Error("Error during message serialization")
			return err
		}
	}

	commandLine := strings.Replace(l.configuration.Command, l.configuration.MessagePlaceholder, msg, -1) //TODO message escaping missing!

	cmd := exec.Command("bash", "-c", commandLine) //TODO lol

	l.logger.WithField("command", cmd.Args).Debug("Command to execute")

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	if l.configuration.EchoOutput {
		fmt.Println(out.String())
	}

	if err != nil {
		if jError := job.RejectMessage(); jError != nil {
			l.logger.WithField("error", jError).Debug("Error on MessageReject")
			return jError
		}
		l.logger.Debug("job rejected")
		return nil //Normal finish of the operation, not an error
	}
	if ackError := job.AckMessage(); ackError != nil {
		l.logger.WithField("error", ackError).Debug("Error on MessageAcknowledge")
		return ackError
	}

	l.logger.Debug("job acknowledged")

	return nil
}

// Configure - configure processor
func (l *Shell) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &l.configuration)
	if l.configuration.MessagePlaceholder == "" {
		l.configuration.MessagePlaceholder = "%msg%"
	}
	l.logger = log.WithFields(log.Fields{
		"type":      "processor",
		"processor": "Shell",
	})
	l.logger.WithField("configuration", l.configuration).Info("Configuration loaded")
	return nil
}
