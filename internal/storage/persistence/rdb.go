package persistence

import (
	"encoding/binary"
	"fmt"
	"ivanSaichkin/myredis/internal/config"
	"ivanSaichkin/myredis/internal/storage"
	"os"
	"sync"
	"time"
)

type RDBManager struct {
	config     *config.PersistanceConfig
	serializer *Serializer
	storage    storage.Storage
	changes    int
	mu         *sync.Mutex
}

func NewRDBManager(config *config.PersistanceConfig, store storage.Storage) *RDBManager {
	return &RDBManager{
		config:     config,
		serializer: NewSerializer(),
		storage:    store,
		changes:    0,
	}
}

func (m *RDBManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tempPath := m.config.GetRDBPath() + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create RDB file: %v", err)
	}
	defer file.Close()

	keys := m.storage.Keys()

	header := RDBFileHeader{
		Magic:     [5]byte{'M', 'Y', 'R', 'D', 'B'},
		Version:   RDBVersion,
		Timestamp: time.Now().Unix(),
		KeyCount:  int32(len(keys)),
	}

	if err := binary.Write(file, binary.BigEndian, header); err != nil {
		return fmt.Errorf("failed write rdb header: %v", err)
	}

	for _, key := range keys {
		value, err := m.storage.Get(key)
		if err != nil {
			continue
		}

		if err := m.writeKey(file, key, value); err != nil {
			return fmt.Errorf("failed to write key %s: %v", key, err)
		}
	}

	if _, err := file.Write([]byte{RDBTypeEOF}); err != nil {
		return fmt.Errorf("failed to write EOF: %v", err)
	}

	checksum := m.calculateChecksum(keys)
	if err := binary.Write(file, binary.BigEndian, checksum); err != nil {
		return fmt.Errorf("failed to write checksum: %v", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync RDB file: %v", err)
	}

	if err := os.Rename(tempPath, m.config.GetRDBPath()); err != nil {
		return fmt.Errorf("failed to rename RDB file: %v", err)
	}

	m.changes = 0
	return nil
}

func (m *RDBManager) writeKey(file *os.File, key string, value *storage.StorageValue) error {
	if !value.ExpiredAt.IsZero() {
		if _, err := file.Write([]byte{RDBTypeExpire}); err != nil {
			return err
		}

		expirationTimestamp := value.ExpiredAt.Unix()
		if err := binary.Write(file, binary.BigEndian, expirationTimestamp); err != nil {
			return err
		}
	}

	rdbType := m.serializer.ConvertToRDB(value.Type)
	if _, err := file.Write([]byte{rdbType}); err != nil {
		return err
	}

	keyData := []byte(key)
	keyLen := m.serializer.EncodeLength(len(keyData))
	if _, err := file.Write(keyLen); err != nil {
		return err
	}
	if _, err := file.Write(keyData); err != nil {
		return err
	}

	valueData, err := m.serializer.SerializeValue(value)
	if err != nil {
		return err
	}

	if _, err := file.Write(valueData); err != nil {
		return err
	}

	return nil
}

func (r *RDBManager) calculateChecksum(keys []string) uint32 {
	var sum uint32
	for _, key := range keys {
		sum += uint32(len(key))
	}
	return sum
}
