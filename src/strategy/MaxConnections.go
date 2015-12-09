package strategy

import (
	"qp"
	"errors"
	"sync"
	"fmt"
	"utils"
)

type MaxConnections struct {
	configuration MaxConnectionsConfiguration
	queue qp.ConsumableQueue
	processor qp.Processor
	process bool
	wait sync.WaitGroup
	jobs chan *qp.SimpleJob
}

type MaxConnectionsConfiguration struct {
	MaxConnections int
	Queue string
	Processor string
}

func (p *MaxConnections) Configure(configuration map[string]interface{}, context *qp.Context) error {

	utils.FillStruct(configuration, &p.configuration)

	if p.configuration.MaxConnections <= 0 {
		panic("MaxConnections option for MaxConnections strategy should be > 0") //PROBABLY SHOULD BE ERROR
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

	return nil
}

func (p *MaxConnections) Start() error {
	if p.process {
		return errors.New("This strategy is already running! You need to Stop() it before calling Start again")
	}

	p.process = true

	p.jobs = make(chan *qp.SimpleJob, p.configuration.MaxConnections)

	go func(){
		//NOT SAFE, probably should implement mutex here on reading of p.process
		for p.process {
			message, _ := p.queue.Consume() //TODO check err
			if (message == nil) {
				fmt.Println("No message, skip processing")
				continue
			}
			job := qp.NewSimpleJob(p.queue, *message)
			p.jobs <- job
		}
		close(p.jobs)
	}()

	process := func (id int, decreaseWaitGroup bool) {
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
	for i := 1; i <= p.configuration.MaxConnections; i++ {
		go process(i, i == 1);
	}

	p.wait.Wait()

	return nil
}

func (p *MaxConnections) Stop() error {
	p.process = false
	p.wait.Wait()
	return nil
}

func (p *MaxConnections) GetStatistics() qp.Statistics {
	return qp.Statistics{}
}