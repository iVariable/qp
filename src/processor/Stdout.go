package processor

import (
	"fmt"
	"qp"
	//	"time"
	//	"math/rand"
)

type Stdout struct {
	configuration qp.Configuration
}

func (l *Stdout) Process(job qp.IJob) error {
	fmt.Printf("[Stdout processor] Received message: %#v\n", job.GetMessage())
	if ackError := job.AckMessage(); ackError != nil {
		panic(ackError.Error()) //TODO what to do?
	}
	return nil
}

func (l *Stdout) Configure(configuration map[string]interface{}) error {
	return nil
}
