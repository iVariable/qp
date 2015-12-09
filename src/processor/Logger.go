package processor
import (
	"qp"
	"fmt"
	"time"
	"math/rand"
)

type Dummy struct {
	configuration qp.Configuration
}

func (l *Dummy) Process(job qp.Job) error {
	fmt.Printf("[Dummy processor] Received message: %s\n", job.GetMessage())
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(1000)) * time.Millisecond) //TODO make configurable
	job.AckMessage()
	return nil
}