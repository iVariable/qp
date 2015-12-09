package main

import (
	"qp"
	"os"
	cont "context"
	"strategy"
	"processor"
	"queue"
	"os/signal"
	"syscall"
	"fmt"
	"time"
)

func main() {

	context := cont.Context{}

	context.Configuration = strategy.MaxConnectionsConfiguration{
		MaxConnections: 2}

	context.Control = make(chan cont.ControlSignal)
	context.IsRunning = false

	handleOsSignals(&context)

	go func(){
		context.SendRun()
	}()

//	go func(){
//		z := time.Tick(5*time.Second)
//		for {
//			<-z
//			if context.IsRunning {
//				context.SendStop()
//			} else {
//				context.SendRun()
//			}
//		}
//	}()

	for signal := range context.Control {
		switch signal.Signal {
		case cont.CONTROL_SIGNAL_RUN:
			fmt.Println("Received RUN signal")
			go run(&context)
		case cont.CONTROL_SIGNAL_STOP:
			fmt.Println("Received STOP signal")
			go stop(&context)
		case cont.CONTROL_SIGNAL_TERMINATE:
			fmt.Println("Received TERMINATE signal")
			os.Exit(signal.ExitCode)
		case cont.CONTROL_SIGNAL_TERMINATE_GRACEFUL:
			fmt.Println("Received TERMINATE_GRACEFUL signal.")
			go func () {
				stop(&context)
				fmt.Println("Normal exit performed")
				context.SendTerminate(0)
			}()
		default:
			panic("Unknown control signal received")
		}
	}
}

func handleOsSignals(context *cont.Context) {
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

func stop(context *cont.Context) {
	context.StrategyInitiatedStop = false //global state
	if err := context.Strategy.Stop(); err != nil {
		panic(err)
	}
	context.IsRunning = false //NOT SAFE
}

func run(context *cont.Context) {
	queue := createQueue(context.Configuration)
	contextor := createProcessor(context.Configuration)
	context.Strategy = createStrategy(context.Configuration)
	context.StrategyInitiatedStop = true

	fmt.Println("Configuring processing strategy")
	context.Strategy.Configure(context.Configuration)

	fmt.Println("Start processing queue")
	context.IsRunning = true //NOT SAFE
	if err := context.Strategy.Start(queue, contextor); err == nil {
		if context.StrategyInitiatedStop {
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
	return &processor.Dummy{}
}

func createStrategy(configuration qp.Configuration) qp.ProcessingStrategy {
	return &strategy.MaxConnections{}
}