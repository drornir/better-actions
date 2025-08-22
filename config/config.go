package config

import "sync"

// Define a struct
type (
	Config struct {
		Log LogConfig `flag:"log" json:"log"`
	}
	LogConfig struct {
		Level  string `flag:"level" json:"level"`
		Format string `flag:"format" json:"format"`
	}
)

var (
	globalConfig Config
	configMutex  sync.RWMutex
)

func SetGlobalConfig(c Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = c
}

func GetConfig() Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}
