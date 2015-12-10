package processor
import (
	"qp"
	"fmt"
//	"time"
//	"math/rand"
)

type Stdout struct {
	configuration qp.Configuration
}

func (l *Stdout) Process(job qp.Job) error {
	fmt.Printf("[Stdout processor] Received message: %#v\n", job.GetMessage())
	job.AckMessage()
	return nil
}

func (l *Stdout) Configure(config interface{}) error {
	return nil
}