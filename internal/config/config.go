package config

import (
	"os"
	"time"
)

type PersistenceConfig struct {
	DataDir      string
	Enabled      bool
	Filename     string
	AutoSave     bool
	SaveInterval time.Duration
}

func DefaultePersistenceConfig() *PersistenceConfig {
	dataDir := "data"
	os.MkdirAll(dataDir, 0755)

	return &PersistenceConfig{
		Enabled:      true,
		DataDir:      dataDir,
		Filename:     "dump.bin",
		AutoSave:     true,
		SaveInterval: 60 * time.Second,
	}
}

type Config struct {
	Address     string
	Persistence PersistenceConfig
}

func DefaulteConfig() *Config {
	return &Config{
		Address:     ":6379",
		Persistence: *DefaultePersistenceConfig(),
	}
}
