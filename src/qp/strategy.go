package qp

import (
	"math/big"
	"time"
)

const StatusStopped = 0
const StatusRunning = 1

type Statistics struct {
	QueueName string
	ProcessedMessages big.Int
	FailedMessaged big.Int
	StartedAt time.Time
	Status int
	MessagesInQueue int
}

type Configuration interface {}

type ProcessingStrategy interface {
	Configure(configuration Configuration) error
	Start(queue ConsumableQueue, processor Processor) error
	Stop() error
	GetStatistics() Statistics
}