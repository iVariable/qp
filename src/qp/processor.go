package qp

type Processor interface {
	Configure(configuration map[string]interface{}) error
	Process(job Job) error
}
