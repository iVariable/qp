package qp

import (
	"math/big"
	"time"
)

const StatusStopped = "Stopped"
const StatusRunning = "Running"

type Statistics struct {
	QueueName         string
	ProcessedMessages big.Int
	FailedMessaged    big.Int
	StartedAt         time.Time
	Status            string
	MessagesInQueue   big.Int
}

type Configuration interface{}

type ProcessingStrategy interface {
	Configure(configuration map[string]interface{}, context *Context) error
	Start() error
	Stop() error
	GetStatistics() Statistics
}
