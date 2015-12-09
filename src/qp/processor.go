package qp

type Processor interface {
	Process(job Job) error
}