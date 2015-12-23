package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"qp"
	"resources"
	"utils"

	"flag"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	logger *log.Entry
	config qp.Config
)

func init() {
	logger = log.WithFields(log.Fields{
		"build": "<buildID>",
	})

	var (
		infoVerbosity  = flag.Bool("v", false, "Overrides log level verbosity to INFO level (default verbosity level is WARN)")
		debugVerbosity = flag.Bool("vv", false, "Overrides log level verbosity to DEBUG level")
		showHelp       = flag.Bool("help", false, "Show this help message")
	)

	flag.Parse()

	if flag.NArg() != 1 || *showHelp {
		fmt.Fprintf(os.Stderr, "Usage of %s: %s [options] path_to_config\n\n", os.Args[0], os.Args[0])
		fmt.Fprintf(os.Stderr, "ARGUMENTS\n")
		fmt.Fprintf(os.Stderr, "  path_to_config\n\tpath to config file\n")
		fmt.Fprintf(os.Stderr, "OPTIONS\n")
		flag.PrintDefaults()
		utils.Quit(utils.ExitCodeOk)
	}

	configFile := flag.Arg(0)

	source, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.WithError(err).Fatal("Can't read config file")
		utils.Quitf(utils.ExitCodeRuntimeError, "Can't read config file: %s", err.Error())
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		logger.WithError(err).Fatal("Can't unmarshal config file")
		utils.Quitf(utils.ExitCodeRuntimeError, "Can't unmarshal config file: %s", err.Error())
	}

	if *infoVerbosity {
		config.General.Log.Level = "info"
	}

	if *debugVerbosity {
		config.General.Log.Level = "debug"
	}
}

func main() {
	context := qp.NewContext(&config)

	load(context)

	go func() {
		context.SendRun()
	}()

	context.DispatchLoop(run, stop, status)
}

func status(context *qp.Context) {
	fmt.Println(context.Strategy.GetStatistics())
}

func load(context *qp.Context) {

	loadLogger(context)

	logger.WithField("config", context.Configuration).Info("Loading main configuration")

	loadQueues(context)
	loadProcessors(context)
	loadStrategies(context)

	context.Strategy = *context.AvailableStrategies[context.Configuration.Strategy[0].Name]

	context.Set("IsRunning", false)
}

func loadLogger(context *qp.Context) {
	level, err := log.ParseLevel(context.Configuration.General.Log.Level)
	if err != nil {
		utils.Quitf(utils.ExitCodeMisconfiguration, "Wrong LogLevel [%s]", context.Configuration.General.Log.Level)
	}
	log.SetLevel(level)
}

func loadQueues(context *qp.Context) {
	for _, config := range context.Configuration.Queue {
		newQueue, ok := resources.AvailableQueues[config.Type]
		if !ok {
			logger.WithField("requestedType", config.Type).Fatal("Unknown queue type requested")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Unknown queue type requested: %s", config.Type)
		}
		newInstance := newQueue()
		if err := newInstance.Configure(config.Options); err != nil {
			logger.WithField("error", err).Fatal("Error configuring queue")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Error configuring queue: %s", err.Error())
		}
		context.AvailableQueues[config.Name] = &newInstance
	}
}

func loadProcessors(context *qp.Context) {
	for _, config := range context.Configuration.Processor {
		newValue, ok := resources.AvailableProcessors[config.Type]
		if !ok {
			logger.WithField("requestedType", config.Type).Fatal("Unknown processor type requested")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Unknown processor type requested: %s", config.Type)
		}
		newInstance := newValue()
		if err := newInstance.Configure(config.Options); err != nil {
			logger.WithField("error", err).Fatal("Error configuring processor")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Error configuring processor: %s", err.Error())
		}
		context.AvailableProcessors[config.Name] = &newInstance
	}
}

func loadStrategies(context *qp.Context) {
	if len(context.Configuration.Strategy) != 1 {
		logger.Fatal("There should be exactly one Strategy configured")
		utils.Quitf(utils.ExitCodeMisconfiguration, "There should be exactly one Strategy configured")
	}
	for _, config := range context.Configuration.Strategy {
		newValue, ok := resources.AvailableStrategies[config.Type]
		if !ok {
			logger.WithField("requestedType", config.Type).Fatal("Unknown strategy type requested")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Unknown strategy type requested: %s", config.Type)
		}
		newInstance := newValue()
		if err := newInstance.Configure(config.Options, context); err != nil {
			logger.WithField("error", err).Fatal("Error configuring strategy")
			utils.Quitf(utils.ExitCodeMisconfiguration, "Error configuring strategy: %s", err.Error())
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

	logger.Info("Configuring processing strategy")
	config := make(map[string]interface{})

	for k, v := range context.Configuration.Strategy[0].Options {
		config[k] = v
	}

	config["Name"] = context.Configuration.Strategy[0].Name

	if err := context.Strategy.Configure(config, context); err != nil {
		logger.WithField("error", err).Fatal("Error configuring strategy")
		utils.Quitf(utils.ExitCodeMisconfiguration, "Error configuring strategy: %s", err.Error())
	}

	logger.Info("Start processing queue")
	context.Set("IsRunning", true)
	if err := context.Strategy.Start(); err == nil {
		if context.GetOrNil("StrategyInitiatedStop").(bool) {
			context.SendTerminate(0)
		}
	} else {
		logger.WithField("error", err).Fatal("Error running strategy")
		utils.Quitf(utils.ExitCodeRuntimeError, "Error running strategy: %s", err.Error())
	}
}
