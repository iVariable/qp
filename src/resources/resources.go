package resources

import (
	"processor"
	"qp"
	"queue"
	"strategy"
)

var AvailableQueues = make(map[string]func() qp.ConsumableQueue)
var AvailableStrategies = make(map[string]func() qp.ProcessingStrategy)
var AvailableProcessors = make(map[string]func() qp.Processor)

func init() {
	//Queues
	AvailableQueues["Dummy"] = func() qp.ConsumableQueue {
		return &queue.Dummy{}
	}
	AvailableQueues["Tail"] = func() qp.ConsumableQueue {
		return &queue.Tail{}
	}
	AvailableQueues["Sqs"] = func() qp.ConsumableQueue {
		return &queue.Sqs{}
	}

	//Strategies
	AvailableStrategies["ParallelProcessing"] = func() qp.ProcessingStrategy {
		return &strategy.ParallelProcessing{}
	}

	//Processors
	AvailableProcessors["Stdout"] = func() qp.Processor {
		return &processor.Stdout{}
	}
	AvailableProcessors["Shell"] = func() qp.Processor {
		return &processor.Shell{}
	}
	AvailableProcessors["HttpProxy"] = func() qp.Processor {
		return &processor.HttpProxy{}
	}
}
