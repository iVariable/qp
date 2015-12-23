package qp

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"utils"
)

// Control signal constants
const (
	ControlSignalRun = 1
	ControlSignalStop = 2
	ControlSignalTerminate = 3
	ControlSignalTerminateGraceful = 4
	ControlSignalStatus = 5
)


type (
	// Config - application configuration
	Config struct {
		General struct {
					Log struct {
							Level string
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

	// Context - application context
	Context struct {
		Configuration Config
		Strategy      IProcessingStrategy

		AvailableQueues     map[string]*IConsumableQueue
		AvailableProcessors map[string]*IProcessor
		AvailableStrategies map[string]*IProcessingStrategy

		control   chan ControlSignal
		data      map[string]interface{}
		dataMutex sync.RWMutex
		logger    *log.Entry
	}

	// ControlSignal - application flow control signal
	ControlSignal struct {
		Signal   int
		ExitCode int
	}
)

// NewContext - constructor for Context
func NewContext(config *Config) *Context {
	context := Context{
		data:                make(map[string]interface{}),
		control:             make(chan ControlSignal),
		AvailableQueues:     make(map[string]*IConsumableQueue),
		AvailableProcessors: make(map[string]*IProcessor),
		AvailableStrategies: make(map[string]*IProcessingStrategy),
		Configuration:       *config,
		logger:              log.WithField("type", "context"),
	}

	if context.Configuration.General.Log.Level == "" {
		context.Configuration.General.Log.Level = "warn"
	}

	return &context
}

// DispatchLoop - runs application dispatch loop
func (c *Context) DispatchLoop(run, stop, status func(c *Context)) {
	c.logger.Debug("Entering DispatchLoop")
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

		go func() {
			for {
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
						c.SendTerminate(utils.ExitCodeShutdownForced)
					case <-signals:
						c.logger.Info("Shutfown forced because of second signal")
						c.SendTerminate(utils.ExitCodeShutdownForced)
					}
				}
			}
		}()
	}()

	for signal := range c.control {
		c.logger.WithField("signal", signal).Debug("Flow signal caught")
		switch signal.Signal {
		case ControlSignalStatus:
			c.logger.Debug("Received STATUS signal")
			go status(c)
		case ControlSignalRun:
			c.logger.Debug("Received RUN signal")
			go run(c)
		case ControlSignalStop:
			c.logger.Debug("Received STOP signal")
			go stop(c)
		case ControlSignalTerminate:
			c.logger.Debug("Received TERMINATE signal")
			utils.Quit(signal.ExitCode)
		case ControlSignalTerminateGraceful:
			c.logger.Debug("Received TERMINATE_GRACEFUL signal.")
			go func() {
				stop(c)
				c.logger.Debug("Normal exit performed")
				c.SendTerminate(utils.ExitCodeOk)
			}()
		default:
			c.logger.Fatal("Unknown control signal received")
			utils.Quitf(utils.ExitCodeRuntimeError, "Unknown control signal received")
		}
	}
}

func (c *Context) sendControlSignal(signal ControlSignal) {
	c.logger.WithField("signal", signal).Debug("Control signal sent")
	c.control <- signal
}

// SendRun - send run control signal
func (c *Context) SendRun() {
	c.sendControlSignal(ControlSignal{Signal: ControlSignalRun})
}

// SendStatus - send status control signal
func (c *Context) SendStatus() {
	c.sendControlSignal(ControlSignal{Signal: ControlSignalStatus})
}

// SendTerminateGraceful - send terminate graceful control signal
func (c *Context) SendTerminateGraceful() {
	c.sendControlSignal(ControlSignal{Signal: ControlSignalTerminateGraceful})
}

// SendStop - send stop control signal
func (c *Context) SendStop() {
	c.sendControlSignal(ControlSignal{Signal: ControlSignalStop})
}

// SendTerminate - send terminate control signal
func (c *Context) SendTerminate(code int) {
	c.sendControlSignal(ControlSignal{Signal: ControlSignalTerminate, ExitCode: code})
}

// Get - thread-safe value getter
func (c *Context) Get(name string) (value interface{}, ok bool) {
	c.dataMutex.RLock()
	value, ok = c.data[name]
	c.dataMutex.RUnlock()
	return
}

// GetOrNil - thread-safe value getter
func (c *Context) GetOrNil(name string) (value interface{}) {
	c.dataMutex.RLock()
	value, _ = c.data[name]
	c.dataMutex.RUnlock()
	return
}

// Set - thread-safe value setter
func (c *Context) Set(name string, value interface{}) {
	c.dataMutex.Lock()
	c.data[name] = value
	c.dataMutex.Unlock()
}

// Unset - thread-safe value unsetter
func (c *Context) Unset(name string) {
	c.dataMutex.Lock()
	delete(c.data, name)
	c.dataMutex.Unlock()
}
