package strategy

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/iVariable/qp/src/qp"
	"github.com/iVariable/qp/src/utils"
	"math/big"
	"sync"
	"time"
)

// OnProcessingError constants
const (
	OnProcessingErrorPanic = "panic"
	OnProcessingErrorWarning = "warning"
	OnProcessingErrorIgnore = "ignore"
)

type (
	// ParallelProcessing - processing strategy
	ParallelProcessing struct {
		configuration parallelProcessingConfiguration
		queue         qp.IConsumableQueue
		processor     qp.IProcessor
		process       bool
		logger        *log.Entry
		stop          chan bool
		wait          sync.WaitGroup
		jobs          chan *qp.SimpleJob
		startedAt     time.Time
	}

	parallelProcessingConfiguration struct {
		Name                string
		MaxThreads          int
		ProcessorThroughput int
		Queue               string
		Processor           string
		OnProcessingError   string
	}

	consumeResult struct {
		message qp.IMessage
		err     error
	}
)

// Configure configures strategy
func (p *ParallelProcessing) Configure(configuration map[string]interface{}, context *qp.Context) error {
	log.WithFields(log.Fields{
		"strategy": "ParallelProcessing",
	}).Debug("Reading configuration")

	utils.FillStruct(configuration, &p.configuration)
	p.logger = log.WithFields(log.Fields{
		"type":     "strategy",
		"strategy": "ParallelProcessing",
		"name":     p.configuration.Name,
	})

	switch p.configuration.OnProcessingError {
	case OnProcessingErrorIgnore:
	case OnProcessingErrorWarning:
	case OnProcessingErrorPanic:
	default:
		panic("Unknown value set for OnProcessingError")
	}

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

	p.logger.WithField("configuration", p.configuration).Info("Configuration loaded")

	return nil
}

// Start starts processing queue
func (p *ParallelProcessing) Start() error {
	p.logger.Info("Start processing")
	if p.process {
		p.logger.Error("Attempt to start already running strategy")
		return errors.New("This strategy is already running! You need to Stop() it before calling Start again")
	}
	p.startedAt = time.Now()
	p.process = true

	p.jobs = make(chan *qp.SimpleJob, p.configuration.MaxThreads)

	//Actual consumer
	go func() {
		p.logger.Debug("Start consuming messages")
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
				p.logger.Info("Recieved stop signal. Stopped messages consuming")
				return
			case message = <-messages:
				messagesProcessed++
				if message.err != nil {
					p.logger.WithField("error", message.err.Error()).Error("Error on message consume")
				} else {
					job := qp.NewSimpleJob(p.queue, message.message)
					p.logger.WithField("message", message.message).Debug("Job created")
					p.jobs <- job
				}
				go consume()
			}
		}
	}()

	process := func(id int, decreaseWaitGroup bool) {
		logger := p.logger.WithField("worker", id)
		logger.Debug("Start worker thread")
		p.wait.Add(1)
		if decreaseWaitGroup {
			p.wait.Done()
		}
		for job := range p.jobs {
			logger.Debug("Recieved job")
			if err := p.processor.Process(job); err != nil {
				switch p.configuration.OnProcessingError {
				case OnProcessingErrorIgnore:
				case OnProcessingErrorWarning:
					logger.WithField("error", err.Error()).Warn("Error on job processing")
				case OnProcessingErrorPanic:
					logger.WithField("error", err.Error()).Fatal("Error on job processing")
					panic("Error while processing: " + err.Error())
				}
			}
		}
		logger.Debug("Worker thread finished")
		p.wait.Done()
	}

	p.logger.Debug("Launching workers")
	p.wait.Add(1)
	for i := 1; i <= p.configuration.MaxThreads; i++ {
		go process(i, i == 1)
	}
	p.wait.Wait()
	p.logger.Debug("All workers finished work")

	return nil
}

// Stop stops queue processing
func (p *ParallelProcessing) Stop() error {
	p.logger.Info("Stopping processing")
	p.process = false
	p.stop <- true
	p.wait.Wait()
	p.logger.Info("Processing stopped")
	return nil
}

// GetStatistics returns stats
func (p *ParallelProcessing) GetStatistics() qp.Statistics {
	var status string
	if p.process {
		status = qp.StatusRunning
	} else {
		status = qp.StatusStopped
	}
	stats := qp.Statistics{
		Status:            status,
		QueueName:         p.configuration.Name,
		ProcessedMessages: *big.NewInt(0),
		FailedMessaged:    *big.NewInt(0),
		StartedAt:         p.startedAt,
		MessagesInQueue:   *big.NewInt(0),
	}
	p.logger.WithFields(log.Fields{
		"Status":            stats.Status,
		"ProcessedMessages": stats.ProcessedMessages,
		"FailedMessaged":    stats.FailedMessaged,
		"StartedAt":         stats.StartedAt,
		"MessagesInQueue":   stats.MessagesInQueue,
	}).Debug("Statistics")
	return stats
}
