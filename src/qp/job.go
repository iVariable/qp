package qp

// IJob job interface
type IJob interface {
	GetMessage() IMessage
	AckMessage() error
	RejectMessage() error
}

// SimpleJob - simple job implementation
type SimpleJob struct {
	queue   IConsumableQueue
	message IMessage
}

// NewSimpleJob Simple job constructor
func NewSimpleJob(q IConsumableQueue, m IMessage) *SimpleJob {
	return &SimpleJob{
		queue:   q,
		message: m}
}

// GetMessage returns message
func (j *SimpleJob) GetMessage() IMessage {
	return j.message
}

// AckMessage acknowledges message
func (j *SimpleJob) AckMessage() error {
	return j.queue.Ack(j.message)
}

// RejectMessage rejects message
func (j *SimpleJob) RejectMessage() error {
	return j.queue.Reject(j.message)
}
