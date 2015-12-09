package context

import "qp"

const CONTROL_SIGNAL_RUN = 1
const CONTROL_SIGNAL_STOP = 2
const CONTROL_SIGNAL_TERMINATE = 3
const CONTROL_SIGNAL_TERMINATE_GRACEFUL = 4

type Context struct {
	Configuration qp.Configuration
	Control chan ControlSignal
	Strategy qp.ProcessingStrategy
	IsRunning bool
	StrategyInitiatedStop bool
}

type ControlSignal struct {
	Signal int
	ExitCode int
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