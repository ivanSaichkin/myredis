package persistence

import (
	"fmt"
	"ivanSaichkin/myredis/internal/config"
	"ivanSaichkin/myredis/internal/storage"
	"sync"
	"time"
)

type PersistenceManager struct {
	config   *config.PersistanceConfig
	storage  storage.Storage
	rdb      *RDBManager
	aof      *AOFManager
	mu       sync.RWMutex
	stopChan chan struct{}
}

func NewPersistenceManager(config *config.PersistanceConfig, store storage.Storage) (*PersistenceManager, error) {
	manager := &PersistenceManager{
		config:   config,
		storage:  store,
		stopChan: make(chan struct{}),
	}

	if config.RDBEnabled {
		manager.rdb = NewRDBManager(config, store)
	}

	if config.AOFEnabled {
		aofManager, err := NewAOFManager(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create AOF manager: %v", err)
		}
		manager.aof = aofManager
	}

	return manager, nil
}

func (p *PersistenceManager) Start() error {
	if !p.config.Enabled {
		return nil
	}

	if p.rdb != nil {
		if err := p.rdb.Load(); err != nil {
			fmt.Printf("Warning: failed to load RDB snapshot: %v\n", err)
		} else {
			fmt.Printf("Loaded RDB snapshot successfully\n")
		}
	}

	if p.aof != nil {
		executor := func(command []string) error {
			fmt.Printf("Replaying AOF command: %v\n", command)
			return nil
		}

		if err := p.aof.Replay(executor); err != nil {
			fmt.Printf("Warning: failed to replay AOF: %v\n", err)
		} else {
			fmt.Printf("Replayed AOF successfully\n")
		}
	}

	if p.rdb != nil {
		p.rdb.StartAutoSave()
	}

	if p.aof != nil {
		go p.startAofRewriting()
	}

	return nil
}

func (p *PersistenceManager) Stop() error {
	close(p.stopChan)

	if p.rdb != nil {
		if err := p.rdb.Save(); err != nil {
			fmt.Printf("Failed to save final RDB snapshot: %v\n", err)
		}
	}

	if p.aof != nil {
		if err := p.aof.Close(); err != nil {
			fmt.Printf("Failed to close AOF: %v\n", err)
		}
	}

	return nil
}

func (p *PersistenceManager) NotifyCommand(command []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.rdb != nil {
		p.rdb.NotifyChanges()
	}

	if p.aof != nil {
		if err := p.aof.Append(command); err != nil {
			fmt.Printf("Failed to append to AOF: %v\n", err)
		}
	}
}

func (p *PersistenceManager) SaveSnapshot() error {
	if p.rdb == nil {
		return fmt.Errorf("RDB persistence is disabled")
	}

	return p.rdb.Save()
}

func (p *PersistenceManager) RewriteAOF() error {
	if p.aof == nil {
		return fmt.Errorf("AOF persistence is disabled")
	}

	snapshotFunc := func() map[string]*storage.StorageValue {
		return p.getStorageSnapshot()
	}

	return p.aof.Rewrite(snapshotFunc)
}

func (p *PersistenceManager) startAofRewriting() {
	if p.aof == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if p.aof.ShouldRewrite() {
				if err := p.RewriteAOF(); err != nil {
					fmt.Printf("AOF rewriting failed: %v\n", err)
				} else {
					fmt.Printf("AOF rewritten successfully\n")
				}
			}
		case <-p.stopChan:
			return
		}
	}
}

func (p *PersistenceManager) getStorageSnapshot() map[string]*storage.StorageValue {
	snapshot := make(map[string]*storage.StorageValue)

	keys := p.storage.Keys()
	for _, key := range keys {
		value, err := p.storage.Get(key)
		if err == nil {
			snapshot[key] = value
		}
	}

	return snapshot
}

func (p *PersistenceManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["enabled"] = p.config.Enabled

	if p.rdb != nil {
		stats["rdb_enabled"] = true
	} else {
		stats["rdb_enabled"] = false
	}

	if p.aof != nil {
		aofStats := p.aof.GetStats()
		for k, v := range aofStats {
			stats["aof_"+k] = v
		}
	} else {
		stats["aof_enabled"] = false
	}

	return stats
}
