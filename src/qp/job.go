package qp

type Job interface {
	GetMessage() Message
	AckMessage() error
	RejectMessage() error
}

type SimpleJob struct {
	queue ConsumableQueue
	message Message
}

func NewSimpleJob(q ConsumableQueue, m Message) *SimpleJob {
	return &SimpleJob{
		queue: q,
		message: m}
}

func (j *SimpleJob) GetMessage() Message {
	return j.message
}

func (j *SimpleJob) AckMessage() error {
	return j.queue.Ack(&j.message)
}

func (j *SimpleJob) RejectMessage() error {
	return j.queue.Reject(&j.message)
}