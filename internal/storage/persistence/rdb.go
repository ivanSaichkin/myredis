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
		KeyCount:  uint32(len(keys)),
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

func (m *RDBManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.config.GetRDBPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open RDB file: %v", err)
	}
	defer file.Close()

	var header RDBFileHeader
	if err := binary.Read(file, binary.BigEndian, &header); err != nil {
		return fmt.Errorf("failed to read RDB header: %v", err)
	}

	if string(header.Magic[:]) != RDBMagicHeader {
		return fmt.Errorf("invalid RDB file format")
	}

	m.storage.Clear()

	for i := uint32(0); i < header.KeyCount; i++ {
		key, value, err := m.readKey(file)
		if err != nil {
			return fmt.Errorf("failed to read key %d: %v", i, err)
		}

		if err := m.storage.Set(key, value); err != nil {
			return fmt.Errorf("failed to restore key %s: %v", key, err)
		}

		if !value.ExpiredAt.IsZero() {
			ttl := time.Until(value.ExpiredAt)
			if ttl > 0 {
				m.storage.Expire(key, ttl)
			}
		}
	}

	var eofMarker byte
	if err := binary.Read(file, binary.BigEndian, &eofMarker); err != nil {
		return fmt.Errorf("failed to read EOF marker: %v", err)
	}
	if eofMarker != RDBTypeEOF {
		return fmt.Errorf("invalid EOF marker")
	}

	var checksum uint32
	if err := binary.Read(file, binary.BigEndian, &checksum); err != nil {
		return fmt.Errorf("failed to read checksum: %v", err)
	}

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

func (m *RDBManager) readKey(file *os.File) (string, *storage.StorageValue, error) {
	var expireAt time.Time

	var typeByte byte
	if err := binary.Read(file, binary.BigEndian, &typeByte); err != nil {
		return "", nil, err
	}

	if typeByte == RDBTypeExpire {
		var expirationTimestamp int64
		if err := binary.Read(file, binary.BigEndian, &expirationTimestamp); err != nil {
			return "", nil, err
		}

		expireAt = time.Unix(expirationTimestamp, 0)
		var typeByte byte
		if err := binary.Read(file, binary.BigEndian, &typeByte); err != nil {
			return "", nil, err
		}
	}

	key, err := m.readString(file)
	if err != nil {
		return "", nil, err
	}

	valueType := m.serializer.ConvertFromRDB(typeByte)
	valueData, err := m.readValueData(file)
	if err != nil {
		return "", nil, err
	}

	valueDataObj, err := m.serializer.DeserializeValue(valueType, valueData)
	if err != nil {
		return "", nil, err
	}

	return key,
		&storage.StorageValue{
			Type:      valueType,
			Data:      valueDataObj,
			ExpiredAt: expireAt,
		},
		nil
}

func (m *RDBManager) readString(file *os.File) (string, error) {
	lenBuf := make([]byte, 1)
	if _, err := file.Read(lenBuf); err != nil {
		return "", err
	}

	length, bytesRead := m.serializer.DecodeLength(lenBuf)
	if bytesRead == 0 {
		return "", fmt.Errorf("invalid string length")
	}

	if bytesRead > 1 {
		additional := make([]byte, bytesRead-1)
		if _, err := file.Read(additional); err != nil {
			return "", nil
		}

		lenBuf = append(lenBuf, additional...)
		length, _ = m.serializer.DecodeLength(lenBuf)
	}

	data := make([]byte, length)
	if _, err := file.Read(data); err != nil {
		return "", err
	}

	return string(data), nil
}

func (m *RDBManager) readValueData(file *os.File) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	currentPos, err := file.Seek(0, 1)
	if err != nil {
		return nil, err
	}

	remaning := stat.Size() - currentPos - 1 - 4
	if remaning <= 0 {
		return nil, fmt.Errorf("invalid value data size")
	}

	data := make([]byte, remaning)
	if _, err := file.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (m *RDBManager) calculateChecksum(keys []string) uint32 {
	var sum uint32
	for _, key := range keys {
		sum += uint32(len(key))
	}
	return sum
}

func (m *RDBManager) NotifyChanges() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.changes++

	if m.config.RDBEnabled && m.changes >= m.config.RDBChangesThreshold {
		go func() {
			if err := m.Save(); err != nil {
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		}()
	}
}

func (m *RDBManager) StartAutoSave() {
	if m.config.RDBEnabled || m.config.RDBSaveInterval == 0 {
		return
	}

	ticker := time.NewTicker(m.config.RDBSaveInterval)
	go func() {
		for range ticker.C {
			if err := m.Save(); err != nil {
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		}
	}()
}
