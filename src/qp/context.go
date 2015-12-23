package qp

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	log "github.com/Sirupsen/logrus"
	"utils"
)

const CONTROL_SIGNAL_RUN = 1
const CONTROL_SIGNAL_STOP = 2
const CONTROL_SIGNAL_TERMINATE = 3
const CONTROL_SIGNAL_TERMINATE_GRACEFUL = 4
const CONTROL_SIGNAL_STATUS = 5

type Config struct {
	General struct {
		Log struct {
			Level string
			Path string
		}
	}
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
	Configuration       Config
	Strategy            IProcessingStrategy

	AvailableQueues     map[string]*IConsumableQueue
	AvailableProcessors map[string]*IProcessor
	AvailableStrategies map[string]*IProcessingStrategy

	control             chan ControlSignal
	data                map[string]interface{}
	dataMutex           sync.RWMutex
	logger				*log.Entry
}

type ControlSignal struct {
	Signal   int
	ExitCode int
}

func NewContext() *Context {
	context := Context{}
	context.data = make(map[string]interface{})
	context.control = make(chan ControlSignal)
	context.AvailableQueues = make(map[string]*IConsumableQueue)
	context.AvailableProcessors = make(map[string]*IProcessor)
	context.AvailableStrategies = make(map[string]*IProcessingStrategy)
	context.logger = log.WithField("type", "context")
	return &context
}

func (c *Context) DispatchLoop(run, stop, status func(c *Context)) {
	c.logger.Debug("Entering DispatchLoop")
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

		go func() {
			for{
				sig := <-signals
				c.logger.WithField("signal", sig).Debug("OS signal caught")
				switch sig {
				case syscall.SIGUSR1:
					c.SendStatus()
				case syscall.SIGINT:
					fallthrough
				case syscall.SIGTERM:
					c.logger.Info("Shutting down gracefully")
					c.SendTerminateGraceful()

					select {
					case <-time.After(30 * time.Second): //TODO make configurable
						c.logger.Info("Shutting down forced after 30 secs")
						c.SendTerminate(utils.EXITCODE_SHUTDOWN_FORCED)
					case <-signals:
						c.logger.Info("Shutfown forced because of second signal")
						c.SendTerminate(utils.EXITCODE_SHUTDOWN_FORCED)
					}
				}
			}
		}()
	}()

	for signal := range c.control {
		c.logger.WithField("signal", signal).Debug("Flow signal caught")
		switch signal.Signal {
		case CONTROL_SIGNAL_STATUS:
			c.logger.Debug("Received STATUS signal")
			go status(c)
		case CONTROL_SIGNAL_RUN:
			c.logger.Debug("Received RUN signal")
			go run(c)
		case CONTROL_SIGNAL_STOP:
			c.logger.Debug("Received STOP signal")
			go stop(c)
		case CONTROL_SIGNAL_TERMINATE:
			c.logger.Debug("Received TERMINATE signal")
			utils.Quit(signal.ExitCode)
		case CONTROL_SIGNAL_TERMINATE_GRACEFUL:
			c.logger.Debug("Received TERMINATE_GRACEFUL signal.")
			go func() {
				stop(c)
				c.logger.Debug("Normal exit performed")
				c.SendTerminate(utils.EXITCODE_OK)
			}()
		default:
			c.logger.Fatal("Unknown control signal received")
			utils.Quitf(utils.EXITCODE_RUNTIME_ERROR, "Unknown control signal received")
		}
	}
}

func (c *Context) sendControlSignal(signal ControlSignal) {
	c.logger.WithField("signal", signal).Debug("Control signal sent")
	c.control <- signal
}

func (c *Context) SendRun() {
	c.sendControlSignal(ControlSignal{Signal: CONTROL_SIGNAL_RUN})
}

func (c *Context) SendStatus() {
	c.sendControlSignal(ControlSignal{Signal: CONTROL_SIGNAL_STATUS})
}

func (c *Context) SendTerminateGraceful() {
	c.sendControlSignal(ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE_GRACEFUL})
}

func (c *Context) SendStop() {
	c.sendControlSignal(ControlSignal{Signal: CONTROL_SIGNAL_STOP})
}

func (c *Context) SendTerminate(code int) {
	c.sendControlSignal(ControlSignal{Signal: CONTROL_SIGNAL_TERMINATE, ExitCode: code})
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
