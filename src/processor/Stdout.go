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
	fmt.Printf("[Stdout processor] Received message: %#s\n", job.GetMessage())
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
//
//	time.Sleep(time.Duration(r.Intn(1000)) * time.Millisecond) //TODO make configurable
	job.AckMessage()
	return nil
}

func (l *Stdout) Configure(config interface{}) error {
	return nil
}