package resources

import (
	"queue"
	"qp"
	"strategy"
	"processor"
)

var AvailableQueues = make(map[string]func()qp.ConsumableQueue)
var AvailableStrategies = make(map[string]func()qp.ProcessingStrategy)
var AvailableProcessors = make(map[string]func()qp.Processor)

func init() {
	//Queues
	AvailableQueues["Dummy"] = func() qp.ConsumableQueue {
		return &queue.Dummy{}
	}

	//Strategies
	AvailableStrategies["MaxConnections"] = func() qp.ProcessingStrategy {
		return &strategy.MaxConnections{}
	}

	//Processors
	AvailableProcessors["Stdout"] = func() qp.Processor {
		return &processor.Stdout{}
	}
}


