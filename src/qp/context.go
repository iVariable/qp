package qp

import (
	"sync"
)

const CONTROL_SIGNAL_RUN = 1
const CONTROL_SIGNAL_STOP = 2
const CONTROL_SIGNAL_TERMINATE = 3
const CONTROL_SIGNAL_TERMINATE_GRACEFUL = 4

type Config struct {
	Strategy []struct{
		Name string
		Type string
		Options map[string]interface{}
	}
	Queue []struct{
		Name string
		Type string
		Options map[string]interface{}
	}
	Processor []struct{
		Name string
		Type string
		Options map[string]interface{}
	}
}

type Context struct {
	Configuration Config
	Strategy ProcessingStrategy

	AvailableQueues map[string]ConsumableQueue
	AvailableProcessors map[string]Processor
	AvailableStrategies map[string]ProcessingStrategy

	Control chan ControlSignal
	data map[string]interface{}
	dataMutex sync.RWMutex
}

type ControlSignal struct {
	Signal int
	ExitCode int
}

func NewContext() *Context {
	context := Context{}
	context.data = make(map[string]interface{})
	context.Control = make(chan ControlSignal)
	context.AvailableQueues = make(map[string]ConsumableQueue)
	context.AvailableProcessors = make(map[string]Processor)
	context.AvailableStrategies = make(map[string]ProcessingStrategy)
	return &context
}

func (c *Context) SendRun() {
	c.Control <- ControlSignal{Signal: CONTROL_SIGNAL_RUN}
}

func (c *Context) SendTerminateGraceful() {
	c.Control <- ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE_GRACEFUL}
}

func (c *Context) SendStop() {
	c.Control <- ControlSignal{Signal: CONTROL_SIGNAL_STOP}
}

func (c *Context) SendTerminate(code int) {
	c.Control <- ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE, ExitCode: code}
}

func (c *Context) Get(name string) (value interface{}, ok bool) {
	c.dataMutex.RLock()
	value, ok = c.data[name]
	c.dataMutex.RUnlock()
	return
}

func (c *Context) GetOrNil(name string) (value interface{}) {
	c.dataMutex.RLock()
	value, _ = c.data[name]
	c.dataMutex.RUnlock()
	return
}

func (c *Context) Set(name string, value interface{}) {
	c.dataMutex.Lock()
	c.data[name] = value
	c.dataMutex.Unlock()
}

func (c *Context) Unset(name string) {
	c.dataMutex.Lock()
	delete(c.data, name)
	c.dataMutex.Unlock()
}