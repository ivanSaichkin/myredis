package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"ivanSaichkin/myredis/internal/config"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PersistenceManager struct {
	config   *config.PersistenceConfig
	storage  Storage
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

func NewPersistenceManager(config *config.PersistenceConfig, store Storage) *PersistenceManager {
	return &PersistenceManager{
		config:   config,
		storage:  store,
		stopChan: make(chan struct{}),
	}
}

func (p *PersistenceManager) Start() error {
	if !p.config.Enabled {
		return nil
	}

	if err := p.Load(); err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}

	if p.config.AutoSave {
		go p.startAutoSave()
	}

	p.running = true
	return nil
}

func (p *PersistenceManager) Stop() error {
	if !p.config.Enabled || !p.running {
		return nil
	}

	close(p.stopChan)

	if err := p.Save(); err != nil {
		return fmt.Errorf("failed to save final snapshot: %v", err)
	}

	p.running = false
	return nil
}

func (p *PersistenceManager) Save() error {
	if !p.config.Enabled {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.storage == nil {
		return fmt.Errorf("storage is not initialized")
	}

	tempPath := p.getFilePath() + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create data file: %v", err)
	}
	defer file.Close()

	keys := p.storage.Keys()

	snapshot := StorageSnapshot{
		Timestamp: time.Now(),
		KeyCount:  len(keys),
		Entries:   make([]StorageEntry, len(keys)),
	}

	for _, key := range keys {
		value, err := p.storage.Get(key)
		if err != nil {
			continue
		}

		var serializableData interface{}
		switch value.Type {
		case StringType:
			serializableData = value.Data
		case HashType:
			if hash, ok := value.Data.(*HashData); ok {
				serializableData = hash.Fields()
			} else {
				fmt.Printf("Warning: invalid hash data for key %s\n", key)
				continue
			}
		case ListType:
			if list, ok := value.Data.(*ListData); ok {
				serializableData = list.GetAll()
			} else {
				fmt.Printf("Warning: invalid list data for key %s\n", key)
				continue
			}
		case SetType:
			if set, ok := value.Data.(*SetData); ok {
				serializableData = set.Members()
			} else {
				fmt.Printf("Warning: invalid set data for key %s\n", key)
				continue
			}
		default:
			serializableData = value.Data
		}

		entry := StorageEntry{
			Key:       key,
			Type:      value.Type,
			Data:      serializableData,
			ExpiredAt: value.ExpiredAt,
		}
		snapshot.Entries = append(snapshot.Entries, entry)
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %v", err)
	}

	dataLen := uint32(len(data))
	if err := binary.Write(file, binary.BigEndian, dataLen); err != nil {
		return fmt.Errorf("failed to write data length: %v", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %v", err)
	}

	if err := os.Rename(tempPath, p.getFilePath()); err != nil {
		return fmt.Errorf("failed to rename data file: %v", err)
	}

	fmt.Printf("Saved %d keys to persistence file\n", len(keys))
	return nil
}

func (p *PersistenceManager) Load() error {
	if !p.config.Enabled {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.storage == nil {
		return fmt.Errorf("storage is not initialized")
	}

	path := p.getFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("No persistence file found, starting with empty storage")
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open data file: %v", err)
	}
	defer file.Close()

	var dataLen uint32
	if err := binary.Read(file, binary.BigEndian, &dataLen); err != nil {
		return fmt.Errorf("failed to read data length: %v", err)
	}

	data := make([]byte, dataLen)
	if _, err := file.Read(data); err != nil {
		return fmt.Errorf("failed to read data: %v", err)
	}

	var snapshot StorageSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot: %v", err)
	}

	p.storage.Clear()

	for _, entry := range snapshot.Entries {
		var data interface{}
		switch entry.Type {
		case StringType:
			if str, ok := entry.Data.(string); ok {
				data = str
			} else {
				fmt.Printf("Warning: invalid string data for key %s: %T\n", entry.Key, entry.Data)
				continue
			}
		case HashType:
			if fields, ok := entry.Data.(map[string]interface{}); ok {
				hash := NewHashData()
				for field, value := range fields {
					if strValue, ok := value.(string); ok {
						hash.Set(field, strValue)
					} else {
						fmt.Printf("Warning: invalid hash field value for key %s.%s: %T\n", entry.Key, field, value)
					}
				}
				data = hash
			} else {
				fmt.Printf("Warning: invalid hash data for key %s: %T\n", entry.Key, entry.Data)
				continue
			}
		case ListType:
			if elements, ok := entry.Data.([]interface{}); ok {
				list := NewListData()
				for _, element := range elements {
					if strElement, ok := element.(string); ok {
						list.PushRight(strElement)
					} else {
						fmt.Printf("Warning: invalid list element for key %s: %T\n", entry.Key, element)
					}
				}
				data = list
			} else {
				fmt.Printf("Warning: invalid list data for key %s: %T\n", entry.Key, entry.Data)
				continue
			}
		case SetType:
			if members, ok := entry.Data.([]interface{}); ok {
				set := NewSetData()
				for _, member := range members {
					if strMember, ok := member.(string); ok {
						set.Add(strMember)
					} else {
						fmt.Printf("Warning: invalid set member for key %s: %T\n", entry.Key, member)
					}
				}
				data = set
			} else {
				fmt.Printf("Warning: invalid set data for key %s: %T\n", entry.Key, entry.Data)
				continue
			}
		default:
			data = entry.Data
		}
		storageValue := &StorageValue{
			Type:      entry.Type,
			Data:      data,
			ExpiredAt: entry.ExpiredAt,
		}

		if err := p.storage.Set(entry.Key, storageValue); err != nil {
			fmt.Printf("Warning: failed to restore key %s: %v\n", entry.Key, err)
			continue
		}
	}

	fmt.Printf("Loaded %d keys from persistence file\n", len(snapshot.Entries))
	return nil
}

func (p *PersistenceManager) getFilePath() string {
	return filepath.Join(p.config.DataDir, p.config.Filename)
}

func (p *PersistenceManager) startAutoSave() {
	ticker := time.NewTicker(p.config.SaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.Save(); err != nil {
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		case <-p.stopChan:
			return
		}
	}
}
