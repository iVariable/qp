package main

import (
	"qp"
	"os"
	"strategy"
	"processor"
	"queue"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"resources"
)

func main() {

	context := qp.NewContext()

	if len(os.Args) != 2 {
		panic("Pls provide config as an argument")
	}

	configFile := os.Args[1]

	load(context, configFile)
	handleOsSignals(context)

	go func(){
		context.SendRun()
	}()

	for signal := range context.Control {
		switch signal.Signal {
		case qp.CONTROL_SIGNAL_RUN:
			fmt.Println("Received RUN signal")
			go run(context)
		case qp.CONTROL_SIGNAL_STOP:
			fmt.Println("Received STOP signal")
			go stop(context)
		case qp.CONTROL_SIGNAL_TERMINATE:
			fmt.Println("Received TERMINATE signal")
			os.Exit(signal.ExitCode)
		case qp.CONTROL_SIGNAL_TERMINATE_GRACEFUL:
			fmt.Println("Received TERMINATE_GRACEFUL signal.")
			go func () {
				stop(context)
				fmt.Println("Normal exit performed")
				context.SendTerminate(0)
			}()
		default:
			panic("Unknown control signal received")
		}
	}
}

func load(context *qp.Context, configFile string) {
	var config qp.Config

	source, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}

	context.Configuration = config
	loadQueues(context)
	loadProcessors(context)
	loadStrategies(context)

	context.Strategy = createStrategy(context.Configuration)

	context.Set("IsRunning", false)
}

func loadQueues(context *qp.Context) {
	for _, config := range context.Configuration.Queue {
		newQueue, ok := resources.AvailableQueues[config.Type]
		if !ok {
			panic(fmt.Sprintf("Unknown queue type requested: %s", config.Type))
		}
		newInstance := newQueue()
		newInstance.Configure(config.Options)
		context.AvailableQueues[config.Name] = &newInstance
	}
}

func loadProcessors(context *qp.Context) {
	for _, config := range context.Configuration.Processor {
		newValue, ok := resources.AvailableProcessors[config.Type]
		if !ok {
			panic(fmt.Sprintf("Unknown processor type requested: %s", config.Type))
		}
		newInstance := newValue()
		newInstance.Configure(config.Options)
		context.AvailableProcessors[config.Name] = &newInstance
	}
}

func loadStrategies(context *qp.Context) {
	if len(context.Configuration.Strategy) != 1 {
		panic("There should be exactly one Strategy configured")
	}
	for _, config := range context.Configuration.Strategy {
		newValue, ok := resources.AvailableStrategies[config.Type]
		if !ok {
			panic(fmt.Sprintf("Unknown strategy type requested: %s", config.Type))
		}
		newInstance := newValue()
		newInstance.Configure(config.Options, context)
		context.AvailableStrategies[config.Name] = &newInstance
	}
}

func handleOsSignals(context *qp.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals

		fmt.Println("Shutting down gracefully")
		context.SendTerminateGraceful()

		select {
		case <- time.After(30 * time.Second): //TODO make configurable
			fmt.Println("Shutting down forced after 30 secs")
			context.SendTerminate(1)
		case <-signals:
			fmt.Println("Shutfown forced because of second signal")
			context.SendTerminate(666)
		}
	}()

}

func stop(context *qp.Context) {
	context.Set("StrategyInitiatedStop", false)
	if err := context.Strategy.Stop(); err != nil {
		panic(err)
	}
	context.Set("IsRunning", false)
}

func run(context *qp.Context) {
	var strategy qp.ProcessingStrategy

	for _, s := range context.AvailableStrategies { //Taking random element from hash map ))
		strategy = *s
		break;
	}

	context.Strategy = strategy
	context.Set("StrategyInitiatedStop", true)

	fmt.Println("Configuring processing strategy")
	if err := strategy.Configure(context.Configuration.Strategy[0].Options, context); err != nil {
		panic("Error configuring strategy: " + err.Error())
	}

	fmt.Println("Start processing queue")
	context.Set("IsRunning", true)
	if err := strategy.Start(); err == nil {
		if context.GetOrNil("StrategyInitiatedStop").(bool) {
			context.SendTerminate(0)
		}
	} else {
		panic(err)
	}
}

func createQueue(configuration qp.Configuration) qp.ConsumableQueue {
	return &queue.Dummy{}
}

func createProcessor(configuration qp.Configuration) qp.Processor {
	return &processor.Stdout{}
}

func createStrategy(configuration qp.Configuration) qp.ProcessingStrategy {
	return &strategy.MaxConnections{}
}