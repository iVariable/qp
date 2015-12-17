package processor

import (
	"fmt"
	"qp"
	"bytes"
	"os/exec"
	"utils"
)

type Shell struct {
	configuration shellConfiguration
}

type shellConfiguration struct {
	Command string
	EchoOutput bool
}

func (l *Shell) Process(job qp.Job) error {
	msg, err := job.GetMessage().Serialize()
	if err != nil {
		panic(err.Error())
	}
	cmd := exec.Command(l.configuration.Command, msg)

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
	return nil
}
