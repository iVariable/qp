package qp

type IProcessor interface {
	Configure(configuration map[string]interface{}) error
	Process(job IJob) error
}
