package persistence

import (
	"bufio"
	"fmt"
	"io"
	"ivanSaichkin/myredis/internal/config"
	"ivanSaichkin/myredis/internal/storage"
	"os"
	"sync"
	"time"
)

type AOFManager struct {
	config     *config.PersistanceConfig
	serializer *Serializer
	file       *os.File
	writer     *bufio.Writer
	mu         sync.RWMutex
	syncTicker *time.Ticker
	stopChan   chan struct{}
}

func NewAOFManager(cfg *config.PersistanceConfig) (*AOFManager, error) {
	manager := &AOFManager{
		config:     cfg,
		serializer: NewSerializer(),
		stopChan:   make(chan struct{}),
	}

	if err := manager.openFile(); err != nil {
		return nil, err
	}

	if manager.config.AOFSyncStrategy == config.SyncEverySec {
		manager.startBackgroundSync()
	}

	return manager, nil
}

func (m *AOFManager) openFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.config.GetAOFPath()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open AOF file: %v", err)
	}

	m.file = file
	m.writer = bufio.NewWriter(file)
	return nil
}

func (a *AOFManager) sync() error {
	if err := a.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush AOF buffer: %v", err)
	}

	if err := a.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync AOF file: %v", err)
	}

	return nil
}

func (a *AOFManager) startBackgroundSync() {
	a.syncTicker = time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-a.syncTicker.C:
				a.mu.Lock()
				a.sync()
				a.mu.Unlock()
			case <-a.stopChan:
				return
			}
		}
	}()
}

func (m *AOFManager) Append(command []string) error {
	if !m.config.AOFEnabled {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := m.serializer.SerializerCommand(command)
	if err != nil {
		return fmt.Errorf("failed to serialize command: %v", err)
	}

	data = append(data, '\n')

	if _, err := m.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to AOF: %v", err)
	}

	switch m.config.AOFSyncStrategy {
	case config.SyncAlways:
		if err := m.sync(); err != nil {
			return err
		}
	case config.SyncEverySec:
		if m.writer.Buffered() > 64*1024 { // 64KB
			if err := m.writer.Flush(); err != nil {
				return fmt.Errorf("failed to flush AOF buffer: %v", err)
			}
		}
	case config.SyncNo:
		if m.writer.Buffered() > 128*1024 { // 128KB
			if err := m.writer.Flush(); err != nil {
				return fmt.Errorf("failed to flush AOF buffer: %v", err)
			}
		}
	}

	return nil
}

func (a *AOFManager) GetBufferStats() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	stats := make(map[string]interface{})
	stats["sync_strategy"] = a.config.AOFSyncStrategy

	if a.writer != nil {
		stats["buffered_bytes"] = a.writer.Buffered()
	}

	if info, err := os.Stat(a.config.GetAOFPath()); err == nil {
		stats["file_size"] = info.Size()
		stats["last_modified"] = info.ModTime()
	}

	return stats
}

func (a *AOFManager) ForceSync() error {
	if !a.config.AOFEnabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	return a.sync()
}

func (a *AOFManager) Replay(executor func(command []string) error) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	path := a.config.GetAOFPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Файл не существует - это нормально
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open AOF file for replay: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	lineNumber := 0

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read AOF file at line %d: %v", lineNumber, err)
		}

		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if len(line) == 0 {
			continue
		}

		command, err := a.serializer.DeserializeCommand(line)
		if err != nil {
			return fmt.Errorf("failed to deserialize command at line %d: %v", lineNumber, err)
		}

		if err := executor(command); err != nil {
			return fmt.Errorf("failed to execute command at line %d: %v", lineNumber, err)
		}

		lineNumber++
	}

	return nil
}

func (a *AOFManager) Rewrite(snapshotFunc func() map[string]*storage.StorageValue) error {
	if !a.config.AOFEnabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	tempPath := a.config.GetAOFPath() + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp AOF file: %v", err)
	}
	defer tempFile.Close()

	tempWriter := bufio.NewWriter(tempFile)

	snapshot := snapshotFunc()

	for key, value := range snapshot {
		if value.IsExpired() {
			continue
		}

		var command []string

		switch value.Type {
		case storage.StringType:
			command = []string{"SET", key, value.Data.(string)}
		case storage.HashType:
			hashData := value.Data.(*storage.HashData)
			fields := hashData.Fields()
			command = []string{"HSET", key}
			for field, fieldValue := range fields {
				command = append(command, field, fieldValue)
			}
		case storage.ListType:
			listData := value.Data.(*storage.ListData)
			elements := listData.GetAll()
			command = []string{"RPUSH", key}
			command = append(command, elements...)
		case storage.SetType:
			setData := value.Data.(*storage.SetData)
			members := setData.Members()
			command = []string{"SADD", key}
			command = append(command, members...)
		}

		data, err := a.serializer.SerializerCommand(command)
		if err != nil {
			return fmt.Errorf("failed to serialize command for key %s: %v", key, err)
		}

		data = append(data, '\n')
		if _, err := tempWriter.Write(data); err != nil {
			return fmt.Errorf("failed to write to temp AOF file: %v", err)
		}

		if !value.ExpiredAt.IsZero() {
			ttl := int(time.Until(value.ExpiredAt).Seconds())
			if ttl > 0 {
				expireCommand := []string{"EXPIRE", key, fmt.Sprintf("%d", ttl)}
				expireData, err := a.serializer.SerializerCommand(expireCommand)
				if err != nil {
					return fmt.Errorf("failed to serialize EXPIRE command: %v", err)
				}
				expireData = append(expireData, '\n')
				if _, err := tempWriter.Write(expireData); err != nil {
					return fmt.Errorf("failed to write EXPIRE command: %v", err)
				}
			}
		}
	}

	if err := tempWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush temp AOF file: %v", err)
	}
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp AOF file: %v", err)
	}

	if a.file != nil {
		a.writer.Flush()
		a.file.Close()
	}

	if err := os.Rename(tempPath, a.config.GetAOFPath()); err != nil {
		return fmt.Errorf("failed to rename AOF file: %v", err)
	}

	if err := a.openFile(); err != nil {
		return fmt.Errorf("failed to reopen AOF file: %v", err)
	}

	return nil
}

func (a *AOFManager) Close() error {
	close(a.stopChan)

	if a.syncTicker != nil {
		a.syncTicker.Stop()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writer != nil {
		a.writer.Flush()
	}

	if a.file != nil {
		return a.file.Close()
	}

	return nil
}

func (a *AOFManager) ShouldRewrite() bool {
	if !a.config.AOFEnabled {
		return false
	}

	info, err := os.Stat(a.config.GetAOFPath())
	if err != nil {
		return false
	}

	return info.Size() >= a.config.AOFRewriteMinSize
}

func (a *AOFManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if !a.config.AOFEnabled {
		stats["enabled"] = false
		return stats
	}

	stats["enabled"] = true
	stats["sync_strategy"] = a.config.AOFSyncStrategy

	info, err := os.Stat(a.config.GetAOFPath())
	if err == nil {
		stats["file_size"] = info.Size()
		stats["last_modified"] = info.ModTime()
	}

	return stats
}
