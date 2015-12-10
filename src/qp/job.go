package qp

type Job interface {
	GetMessage() IMessage
	AckMessage() error
	RejectMessage() error
}

type SimpleJob struct {
	queue ConsumableQueue
	message IMessage
}

func NewSimpleJob(q ConsumableQueue, m IMessage) *SimpleJob {
	return &SimpleJob{
		queue: q,
		message: m}
}

func (j *SimpleJob) GetMessage() IMessage {
	return j.message
}

func (j *SimpleJob) AckMessage() error {
	return j.queue.Ack(&j.message)
}

func (j *SimpleJob) RejectMessage() error {
	return j.queue.Reject(&j.message)
}