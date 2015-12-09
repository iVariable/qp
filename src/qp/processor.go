package qp

type Processor interface {
	Configure(config interface{}) error
	Process(job Job) error
}