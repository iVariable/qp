package strategy

import (
	"errors"
	"fmt"
	"qp"
	"sync"
	"utils"
	"time"
	"math/big"
)

type ParallelProcessing struct {
	configuration parallelProcessingConfiguration
	queue         qp.IConsumableQueue
	processor     qp.IProcessor
	process       bool
	stop          chan bool
	wait          sync.WaitGroup
	jobs          chan *qp.SimpleJob
	startedAt	  time.Time
}

type parallelProcessingConfiguration struct {
	Name string
	MaxThreads int
	ProcessorThroughput int
	Queue      string
	Processor  string
}

type consumeResult struct {
	message qp.IMessage
	err     error
}

func (p *ParallelProcessing) Configure(configuration map[string]interface{}, context *qp.Context) error {

	utils.FillStruct(configuration, &p.configuration)

	if p.configuration.MaxThreads <= 0 {
		panic("MaxThreads option for ParallelProcessing strategy should be > 0") //PROBABLY SHOULD BE ERROR
	}

	if queue, ok := context.AvailableQueues[p.configuration.Queue]; !ok {
		panic("Unknown Queue requested")
	} else {
		p.queue = *queue
	}

	if processor, ok := context.AvailableProcessors[p.configuration.Processor]; !ok {
		panic("Unknown Queue requested")
	} else {
		p.processor = *processor
	}

	p.stop = make(chan bool)

	return nil
}

func (p *ParallelProcessing) Start() error {
	if p.process {
		return errors.New("This strategy is already running! You need to Stop() it before calling Start again")
	}
	p.startedAt = time.Now()
	p.process   = true

	p.jobs = make(chan *qp.SimpleJob, p.configuration.MaxThreads)

	//Actual consumer
	go func() {
		messages := make(chan *consumeResult)
		consume := func() {
			message, err := p.queue.Consume()
			messages <- &consumeResult{message, err}
		}

		var message *consumeResult

		go consume()

		prevSecond := time.Now().Second()
		messagesProcessed := 0

		for {
			if p.configuration.ProcessorThroughput > 0 {
				for messagesProcessed >= p.configuration.ProcessorThroughput { //can be done better with time.After(), etc
					if prevSecond != time.Now().Second() {
						messagesProcessed = 0
						prevSecond = time.Now().Second()
					} else {
						time.Sleep(1 * time.Millisecond)
					}
				}
			}

			select {
			case <-p.stop:
				close(p.jobs)
				return
			case message = <-messages:
				messagesProcessed++
				if message.err != nil {
					fmt.Println("Error on message consume: " + message.err.Error())
				} else {
					job := qp.NewSimpleJob(p.queue, message.message)
					p.jobs <- job

				}
				go consume()
			}
		}
	}()

	process := func(id int, decreaseWaitGroup bool) {
		p.wait.Add(1)
		if decreaseWaitGroup {
			p.wait.Done()
		}
		for job := range p.jobs {
			fmt.Println(fmt.Sprintf("[Worker %v] Processing job", id))
			p.processor.Process(job)
		}
		p.wait.Done()
	}

	p.wait.Add(1)
	for i := 1; i <= p.configuration.MaxThreads; i++ {
		go process(i, i == 1)
	}

	p.wait.Wait()

	return nil
}

func (p *ParallelProcessing) Stop() error {
	p.process = false
	p.stop <- true
	p.wait.Wait()
	return nil
}

func (p *ParallelProcessing) GetStatistics() qp.Statistics {
	var status string
	if p.process {
		status = qp.StatusRunning
	} else {
		status = qp.StatusStopped
	}
	return qp.Statistics{
		Status: status,
		QueueName: p.configuration.Name,
		ProcessedMessages: *big.NewInt(0),
		FailedMessaged: *big.NewInt(0),
		StartedAt: p.startedAt,
		MessagesInQueue: *big.NewInt(0),
	}
}
