package resources

import (
	"processor"
	"qp"
	"queue"
	"strategy"
)

var AvailableQueues = make(map[string]func() qp.IConsumableQueue)
var AvailableStrategies = make(map[string]func() qp.IProcessingStrategy)
var AvailableProcessors = make(map[string]func() qp.IProcessor)

func init() {
	//Queues
	AvailableQueues["Dummy"] = func() qp.IConsumableQueue {
		return &queue.Dummy{}
	}
	AvailableQueues["Tail"] = func() qp.IConsumableQueue {
		return &queue.Tail{}
	}
	AvailableQueues["Sqs"] = func() qp.IConsumableQueue {
		return &queue.Sqs{}
	}

	//Strategies
	AvailableStrategies["ParallelProcessing"] = func() qp.IProcessingStrategy {
		return &strategy.ParallelProcessing{}
	}

	//Processors
	AvailableProcessors["Stdout"] = func() qp.IProcessor {
		return &processor.Stdout{}
	}
	AvailableProcessors["Shell"] = func() qp.IProcessor {
		return &processor.Shell{}
	}
	AvailableProcessors["HttpProxy"] = func() qp.IProcessor {
		return &processor.HttpProxy{}
	}
}
