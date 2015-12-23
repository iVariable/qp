package qp

import (
	"math/big"
	"time"
)

// Processing strategy status constants
const (
	StatusStopped = "Stopped"
	StatusRunning = "Running"
)

// Statistics - Processing strategy statistics
type Statistics struct {
	QueueName         string
	ProcessedMessages big.Int
	FailedMessaged    big.Int
	StartedAt         time.Time
	Status            string
	MessagesInQueue   big.Int
}

// IProcessingStrategy processing strategy interface
type IProcessingStrategy interface {
	Configure(configuration map[string]interface{}, context *Context) error
	Start() error
	Stop() error
	GetStatistics() Statistics
}
