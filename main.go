package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/parashmaity/fleare/internal/logger"

	"github.com/parashmaity/fleare/config"

	commander "github.com/parashmaity/fleare/commander"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func main() {

	configFileContent := flag.String("config", "", "Config file content")
	logLevel := flag.String("loglevel", "", "Set log level (debug, info, warn, error)")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	shortVersionFlag := flag.Bool("v", false, "Print the version and exit (shorthand)")

	// Parse the command-line flags
	flag.Parse()

	defaultConfig := config.LoadDefaultConfigYAML()

	// Load override config from flag content
	overrideConfig, err := LoadConfigFromString(*configFileContent)
	if err != nil {
		panic("Error loading override config " + err.Error()) // Failing fast
	}

	// Merge configurations
	config.MergeConfigs(defaultConfig, overrideConfig)

	if logLevel == nil || *logLevel == "" {
		logLevel = &config.GetConfig().Logging.Level
	}
	// Initialize logger
	logger.InitLogger(config.GetConfig().Logging.File, logLevel)

	// If either version flag is set, print the version and exit
	if *versionFlag || *shortVersionFlag {
		fmt.Printf("* %s\n* Version: %s\n* Build Date: %s\n", commander.ProjectName, commander.Version, commander.BuildDate)
		return
	}

	logger.Console("-> Configuration setup successful", zerolog.InfoLevel, nil)

	commander.Execute()

	// Gracefully shutdown logger (flush pending logs)
	logger.CloseLogger()
}

// LoadConfigFromString parses YAML from a string
func LoadConfigFromString(yamlContent string) (*config.Configuration, error) {
	config := &config.Configuration{}
	if yamlContent != "" {
		if err := yaml.Unmarshal([]byte(yamlContent), config); err != nil {
			return nil, err
		}
		return config, nil
	}

	configEnv := os.Getenv("config")
	if configEnv != "" {
		if err := yaml.Unmarshal([]byte(configEnv), config); err != nil {
			return nil, err
		}
		return config, nil
	}

	return config, nil
}
