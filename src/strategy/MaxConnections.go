package strategy

import (
	"qp"
	"errors"
	"sync"
	"fmt"
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
}

func (p *MaxConnections) Configure(configuration qp.Configuration) error {
	if config, ok := configuration.(MaxConnectionsConfiguration); ok {
		p.configuration = config
	} else {
		return errors.New("Wrong configuration provided")
	}

	return nil
}

func (p *MaxConnections) Start(queue qp.ConsumableQueue, processor qp.Processor) error {
	if p.process {
		return errors.New("This strategy is already running! You need to Stop() it before calling Start again")
	}

	p.processor = processor
	p.queue = queue
	p.process = true

	p.jobs = make(chan *qp.SimpleJob, p.configuration.MaxConnections)

	go func(){
		//NOT SAFE, probably should implement mutex here on reading of p.process
		for p.process {
			message, _ := queue.Consume()
			if (message == nil) {
				fmt.Println("No message, skip processing")
				continue
			}
			job := qp.NewSimpleJob(queue, *message)
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
			processor.Process(job)
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