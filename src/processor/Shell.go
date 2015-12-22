package processor

import (
	"fmt"
	"qp"
	"bytes"
	"os/exec"
	"utils"
	"strings"
	log "github.com/Sirupsen/logrus"
)

type Shell struct {
	configuration shellConfiguration
	logger *log.Entry
}

type shellConfiguration struct {
	Command string
	MessagePlaceholder string
	EchoOutput bool
}

func (l *Shell) Process(job qp.IJob) error {
	l.logger.WithField("job", job).Debug("Processing job")
	msg, err := job.GetMessage().Serialize()
	if err != nil {
		l.logger.WithField("error", err).Error("Error during message serialization")
		return err
	}

	commandLine := strings.Replace(l.configuration.Command, l.configuration.MessagePlaceholder, msg, -1)

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

func (l *Shell) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &l.configuration)
	if l.configuration.MessagePlaceholder == "" {
		l.configuration.MessagePlaceholder = "%msg%"
	}
	l.logger = log.WithFields(log.Fields{
		"type": "processor",
		"processor": "Shell",
	})
	l.logger.WithFields(log.Fields{
		"Command": l.configuration.Command,
		"MessagePlaceholder": l.configuration.MessagePlaceholder,
		"EchoOutput": l.configuration.EchoOutput,
	}).Info("Configuration loaded")
	return nil
}
