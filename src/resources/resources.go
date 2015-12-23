package resources

import (
	"processor"
	"qp"
	"queue"
	"strategy"
)

var (
	// AvailableQueues list of configured ready-to-use queues
	AvailableQueues = make(map[string]func() qp.IConsumableQueue)

	// AvailableStrategies list of configured ready-to-use processing strategies
	AvailableStrategies = make(map[string]func() qp.IProcessingStrategy)

	// AvailableProcessors list of configured ready-to-use processors
	AvailableProcessors = make(map[string]func() qp.IProcessor)
)

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
	AvailableProcessors["HTTPProxy"] = func() qp.IProcessor {
		return &processor.HTTPProxy{}
	}
}
