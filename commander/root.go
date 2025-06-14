package commander

import (
	"fmt"
	"os"

	_ "github.com/parashmaity/fleare/commander/list_cmd"
	_ "github.com/parashmaity/fleare/commander/map_cmd"
	_ "github.com/parashmaity/fleare/commander/num_cmd"
	_ "github.com/parashmaity/fleare/commander/string_cmd"
	"github.com/parashmaity/fleare/config"
	"github.com/parashmaity/fleare/internal/logger"
	"github.com/parashmaity/fleare/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configFileContent string
var configFilePath string
var logLevel string

var RootCmd = &cobra.Command{
	Use:   "Fleare",
	Short: "An in-memory database",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Apply config file path
		if configFilePath != "" {
			config.SetConfigPath(configFilePath)
		}
		defaultConfig := config.LoadDefaultConfigYAML()

		// Load override config from --config-content
		overrideConfig, err := LoadConfigFromString(configFileContent)
		if err != nil {
			panic("Error loading override config: " + err.Error()) // Fail fast
		}

		// Merge
		config.MergeConfigs(defaultConfig, overrideConfig)

		if logLevel == "" {
			logLevel = config.GetConfig().Logging.Level
		}

		// Initialize logger
		logger.InitLogger(config.GetConfig().Logging.File, &logLevel)

		logger.Console("-> Configuration setup successful", zerolog.InfoLevel, nil)
	},
	Run: func(cmd *cobra.Command, args []string) {
		server.Start()

		// Gracefully shutdown logger (flush pending logs)
		logger.CloseLogger()
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&configFileContent, "config", "", "Config file content (YAML string)")
	RootCmd.PersistentFlags().StringVar(&configFilePath, "config-path", "", "Path to config file")
	RootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "", "Set log level (debug, info, warn, error)")
	RootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version and exit")
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetBool("version")
		if v {
			cmd.Printf("* %s\n* Version: %s\n* Build Date: %s\n", ProjectName, Version, BuildDate)
			os.Exit(0)
		}
	}
}

func Execute() {
	// Execute the root command
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
