package qp

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const CONTROL_SIGNAL_RUN = 1
const CONTROL_SIGNAL_STOP = 2
const CONTROL_SIGNAL_TERMINATE = 3
const CONTROL_SIGNAL_TERMINATE_GRACEFUL = 4

type Config struct {
	Strategy []struct {
		Name    string
		Type    string
		Options map[string]interface{}
	}
	Queue []struct {
		Name    string
		Type    string
		Options map[string]interface{}
	}
	Processor []struct {
		Name    string
		Type    string
		Options map[string]interface{}
	}
}

type Context struct {
	Configuration Config
	Strategy      ProcessingStrategy

	AvailableQueues     map[string]*ConsumableQueue
	AvailableProcessors map[string]*Processor
	AvailableStrategies map[string]*ProcessingStrategy

	control   chan ControlSignal
	data      map[string]interface{}
	dataMutex sync.RWMutex
}

type ControlSignal struct {
	Signal   int
	ExitCode int
}

func NewContext() *Context {
	context := Context{}
	context.data = make(map[string]interface{})
	context.control = make(chan ControlSignal)
	context.AvailableQueues = make(map[string]*ConsumableQueue)
	context.AvailableProcessors = make(map[string]*Processor)
	context.AvailableStrategies = make(map[string]*ProcessingStrategy)
	return &context
}

func (c *Context) DispatchLoop(run, stop func(c *Context)) {

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-signals

			fmt.Println("Shutting down gracefully")
			c.SendTerminateGraceful()

			select {
			case <-time.After(30 * time.Second): //TODO make configurable
				fmt.Println("Shutting down forced after 30 secs")
				c.SendTerminate(1)
			case <-signals:
				fmt.Println("Shutfown forced because of second signal")
				c.SendTerminate(666)
			}
		}()
	}()

	for signal := range c.control {
		switch signal.Signal {
		case CONTROL_SIGNAL_RUN:
			fmt.Println("Received RUN signal")
			go run(c)
		case CONTROL_SIGNAL_STOP:
			fmt.Println("Received STOP signal")
			go stop(c)
		case CONTROL_SIGNAL_TERMINATE:
			fmt.Println("Received TERMINATE signal")
			os.Exit(signal.ExitCode)
		case CONTROL_SIGNAL_TERMINATE_GRACEFUL:
			fmt.Println("Received TERMINATE_GRACEFUL signal.")
			go func() {
				stop(c)
				fmt.Println("Normal exit performed")
				c.SendTerminate(0)
			}()
		default:
			panic("Unknown control signal received")
		}
	}
}

func (c *Context) SendRun() {
	c.control <- ControlSignal{Signal: CONTROL_SIGNAL_RUN}
}

func (c *Context) SendTerminateGraceful() {
	c.control <- ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE_GRACEFUL}
}

func (c *Context) SendStop() {
	c.control <- ControlSignal{Signal: CONTROL_SIGNAL_STOP}
}

func (c *Context) SendTerminate(code int) {
	c.control <- ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE, ExitCode: code}
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
