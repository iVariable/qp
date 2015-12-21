package processor

import (
	"fmt"
	"qp"
	"bytes"
	"os/exec"
	"utils"
	"strings"
)

type Shell struct {
	configuration shellConfiguration
}

type shellConfiguration struct {
	Command string
	MessagePlaceholder string
	EchoOutput bool
}

func (l *Shell) Process(job qp.Job) error {
	msg, err := job.GetMessage().Serialize()
	if err != nil {
		panic(err.Error())
	}

	commandLine := strings.Replace(l.configuration.Command, l.configuration.MessagePlaceholder, msg, -1)

	cmd := exec.Command("bash", "-c", commandLine) //lol

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		if jError := job.RejectMessage(); jError != nil {
			panic(jError.Error()) //TODO what to do?
		}
		return err
	}
	if ackError := job.AckMessage(); ackError != nil {
		panic(ackError.Error()) //TODO what to do?
	}

	if l.configuration.EchoOutput {
		fmt.Println(out.String())
	}

	return nil
}

func (l *Shell) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &l.configuration)
	if l.configuration.MessagePlaceholder == "" {
		l.configuration.MessagePlaceholder = "%msg%"
	}
	return nil
}
