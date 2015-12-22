package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"qp"
	"resources"

	"gopkg.in/yaml.v2"
	log "github.com/Sirupsen/logrus"
)

func main() {

	context := qp.NewContext()

	if len(os.Args) != 2 {
		panic("Pls provide config as an argument")
	}

	configFile := os.Args[1]

	load(context, configFile)

	go func() {
		context.SendRun()
	}()

	context.DispatchLoop(run, stop, status)
}

func status(context *qp.Context) {
	fmt.Println(context.Strategy.GetStatistics())
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

	loadLogger(context)
	loadQueues(context)
	loadProcessors(context)
	loadStrategies(context)

	context.Strategy = *context.AvailableStrategies[context.Configuration.Strategy[0].Name]

	context.Set("IsRunning", false)
}

func loadLogger(context *qp.Context) {
	level, err := log.ParseLevel(context.Configuration.General.Log.Level)
	if err != nil {
		panic("Wrong LogLevel ["+context.Configuration.General.Log.Level+"]!")
	}
	log.SetLevel(level)
}

func loadQueues(context *qp.Context) {
	for _, config := range context.Configuration.Queue {
		newQueue, ok := resources.AvailableQueues[config.Type]
		if !ok {
			panic(fmt.Sprintf("Unknown queue type requested: %s", config.Type))
		}
		newInstance := newQueue()
		if err := newInstance.Configure(config.Options); err != nil {
			panic("Error configuring queue: " + err.Error())
		}
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
		if err := newInstance.Configure(config.Options); err != nil {
			panic("Error configuring processor: " + err.Error())
		}
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
		if err := newInstance.Configure(config.Options, context); err != nil {
			panic("Error configuring strategy: " + err.Error())
		}
		context.AvailableStrategies[config.Name] = &newInstance
	}
}

func stop(context *qp.Context) {
	context.Set("StrategyInitiatedStop", false)
	if err := context.Strategy.Stop(); err != nil {
		panic(err)
	}
	context.Set("IsRunning", false)
}

func run(context *qp.Context) {
	context.Set("StrategyInitiatedStop", true)

	fmt.Println("Configuring processing strategy")
	config := make(map[string]interface{})

	for k,v := range context.Configuration.Strategy[0].Options {
		config[k] = v
	}

	config["Name"] = context.Configuration.Strategy[0].Name

	if err := context.Strategy.Configure(config, context); err != nil {
		panic("Error configuring strategy: " + err.Error())
	}

	fmt.Println("Start processing queue")
	context.Set("IsRunning", true)
	if err := context.Strategy.Start(); err == nil {
		if context.GetOrNil("StrategyInitiatedStop").(bool) {
			context.SendTerminate(0)
		}
	} else {
		panic(err)
	}
}
