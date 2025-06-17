package config

import (
	"log"
	"os"
	"runtime"

	"github.com/parashmaity/fleare/internal/auth"
	"github.com/parashmaity/fleare/internal/logger"

	"gopkg.in/yaml.v3"
)

const (
	APP_NAME       = "Fleare"
	STATUS_SUCCESS = "Ok"
	STATUS_ERROR   = "Error"
)

var (
	CONFIG_PATH_MAC_OS  = "/etc/fleare/config.yaml"
	CONFIG_PATH_LINUX   = "/etc/fleare/config.yaml"
	CONFIG_PATH_WINDOWS = "/etc/fleare/config.yaml"
)

// OS Enum using iota
type OS int

const (
	Unknown OS = iota
	MacOS
	Linux
	Windows
)

var configInstance *Configuration

type Configuration struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`

	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`

	Memory struct {
		MaxSizeMB      int    `yaml:"max_size_mb"`
		EvictionPolicy string `yaml:"eviction_policy"`
	} `yaml:"memory"`

	Security struct {
		EnableAuth bool   `yaml:"enable_auth"`
		AuthMethod string `yaml:"auth_method"`
		Users      []struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Role     string `yaml:"role"`
		} `yaml:"users"`
	} `yaml:"security"`

	Persistence struct {
		Enable          bool   `yaml:"enable"`
		Path            string `yaml:"path"`
		AfterWriteCount int    `yaml:"after_write_count"`
	} `yaml:"persistence"`

	Shard struct {
		Mode       string `yaml:"mode"`
		ShardCount int    `yaml:"shard_count"`
	} `yaml:"shard"`

	Backup struct {
		Enable          bool   `yaml:"enable"`
		IntervalMinutes int    `yaml:"interval_minutes"`
		BackupDir       string `yaml:"backup_dir"`
	} `yaml:"backup"`

	Misc struct {
		MaxConnections int  `yaml:"max_connections"`
		TimeoutSeconds int  `yaml:"timeout_seconds"`
		StrictMode     bool `yaml:"insertmode"`
	} `yaml:"misc"`
}

func (os OS) String() string {
	switch os {
	case MacOS:
		return "macOS 🍏"
	case Linux:
		return "Linux 🐧"
	case Windows:
		return "Windows 🏁"
	default:
		return "Unknown OS ❓"
	}
}

// Function to get the OS Enum
func GetOS() OS {

	switch runtime.GOOS {
	case "darwin":
		return MacOS
	case "linux":
		return Linux
	case "windows":
		return Windows
	default:
		return Unknown
	}
}

func GetConfigPath() string {

	currentOn := GetOS()
	// fmt.Println("Fleare currently running on:", currentOn.String())
	switch currentOn {
	case MacOS:
		return CONFIG_PATH_MAC_OS
	case Linux:
		return CONFIG_PATH_LINUX
	case Windows:
		return CONFIG_PATH_WINDOWS
	default:
		return CONFIG_PATH_LINUX
	}
}
func SetConfigPath(path string) {
	CONFIG_PATH_MAC_OS, CONFIG_PATH_LINUX, CONFIG_PATH_WINDOWS = path, path, path
}

// Load JSON Config
func LoadDefaultConfigYAML() *Configuration {
	if configInstance == nil {

		file, err := os.ReadFile(GetConfigPath())
		if err != nil {
			logger.Error("Failed to load Configuration file", err, map[string]any{
				"filepath": GetConfigPath(),
			})
			log.Fatal(err)
		}
		err = yaml.Unmarshal(file, &configInstance)
		if err != nil {
			logger.Error("Failed to load Configuration file file format error", err, map[string]any{
				"filepath": GetConfigPath(),
			})
			log.Fatal(err)
		}
		return configInstance
	} else {
		return configInstance
	}
}

func GetConfig() *Configuration {
	if configInstance == nil {
		return LoadDefaultConfigYAML()
	}
	return configInstance
}

// MergeConfigs merges an override config into a default config
func MergeConfigs(defaultConfig, overrideConfig *Configuration) *Configuration {
	// Merge server settings
	if overrideConfig.Server.Host != "" {
		defaultConfig.Server.Host = overrideConfig.Server.Host
	}
	if overrideConfig.Server.Port != 0 {
		defaultConfig.Server.Port = overrideConfig.Server.Port
	}

	// Merge logging settings
	if overrideConfig.Logging.Level != "" {
		defaultConfig.Logging.Level = overrideConfig.Logging.Level
	}
	if overrideConfig.Logging.File != "" {
		defaultConfig.Logging.File = overrideConfig.Logging.File
	}

	// Merge memory settings
	if overrideConfig.Memory.MaxSizeMB != 0 {
		defaultConfig.Memory.MaxSizeMB = overrideConfig.Memory.MaxSizeMB
	}
	if overrideConfig.Memory.EvictionPolicy != "" {
		defaultConfig.Memory.EvictionPolicy = overrideConfig.Memory.EvictionPolicy
	}

	// Merge security settings
	if overrideConfig.Security.EnableAuth {
		defaultConfig.Security.EnableAuth = overrideConfig.Security.EnableAuth
	}
	if overrideConfig.Security.AuthMethod != "" {
		defaultConfig.Security.AuthMethod = overrideConfig.Security.AuthMethod
	}
	if len(overrideConfig.Security.Users) > 0 {
		defaultConfig.Security.Users = overrideConfig.Security.Users
	}

	var isRootFound = false
	for _, user := range defaultConfig.Security.Users {
		if user.Username == "root" {
			isRootFound = true
		}
		auth.UserStore.Add(user.Username, user.Password, user.Role)
	}

	if !isRootFound {
		log.Fatal("Configuration Error, root user not found!. Plese add root user in the Configuration file")
	}

	// Merge persistence settings
	if overrideConfig.Persistence.Enable {
		defaultConfig.Persistence.Enable = overrideConfig.Persistence.Enable
	}
	if overrideConfig.Persistence.Path != "" {
		defaultConfig.Persistence.Path = overrideConfig.Persistence.Path
	}
	if overrideConfig.Persistence.AfterWriteCount != 0 {
		defaultConfig.Persistence.AfterWriteCount = overrideConfig.Persistence.AfterWriteCount
	}

	// Merge backup settings
	if overrideConfig.Backup.Enable {
		defaultConfig.Backup.Enable = overrideConfig.Backup.Enable
	}
	if overrideConfig.Backup.IntervalMinutes != 0 {
		defaultConfig.Backup.IntervalMinutes = overrideConfig.Backup.IntervalMinutes
	}
	if overrideConfig.Backup.BackupDir != "" {
		defaultConfig.Backup.BackupDir = overrideConfig.Backup.BackupDir
	}

	// Merge misc settings
	if overrideConfig.Misc.MaxConnections != 0 {
		defaultConfig.Misc.MaxConnections = overrideConfig.Misc.MaxConnections
	}
	if overrideConfig.Misc.TimeoutSeconds != 0 {
		defaultConfig.Misc.TimeoutSeconds = overrideConfig.Misc.TimeoutSeconds
	}
	defaultConfig.Misc.StrictMode = overrideConfig.Misc.StrictMode

	return defaultConfig
}
