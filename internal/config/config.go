package config

import (
	"os"
	"path/filepath"
	"time"
)

type SyncStrategy string

const (
	SyncAlways   SyncStrategy = "always"
	SyncEverySec SyncStrategy = "everysec"
	SyncNo       SyncStrategy = "no"
)

type PersistanceConfig struct {
	DataDir string
	Enabled bool

	//RDB
	RDBEnabled          bool
	RDBSaveInterval     time.Duration
	RDBChangesThreshold int
	RDBFileName         string

	//AOF
	AOFEnabled        bool
	AOFFilename       string
	AOFSyncStrategy   SyncStrategy
	AOFRewriteMinSize int64
}

type Config struct {
	Address     string
	Persistance PersistanceConfig
}

func DefaulteConfig() *Config {
	dataDir := "data"
	os.MkdirAll(dataDir, 0755)

	return &Config{
		Address: ":6379",
		Persistance: PersistanceConfig{
			Enabled: true,
			DataDir: dataDir,

			RDBEnabled:          true,
			RDBSaveInterval:     60 * time.Second,
			RDBChangesThreshold: 1000,
			RDBFileName:         "dump.rdb",

			AOFEnabled:        false,
			AOFFilename:       "appendonly.aof",
			AOFSyncStrategy:   SyncEverySec,
			AOFRewriteMinSize: 64 * 1024 * 1024,
		},
	}
}

func (c *Config) GetRDBPath() string {
	return filepath.Join(c.Persistance.DataDir, c.Persistance.RDBFileName)
}

func (c *Config) GetAOFPath() string {
	return filepath.Join(c.Persistance.DataDir, c.Persistance.AOFFilename)
}
