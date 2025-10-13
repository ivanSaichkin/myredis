package persistence

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"ivanSaichkin/myredis/internal/storage"
	"time"
)

const (
	RDBMagicHeader = "MYRDB"
	RDBVersion     = 1
	RDBDelimiter   = 0xFF
)

const (
	RDBTypeString = 0
	RDBTypeHash   = 1
	RDBTypeList   = 2
	RDBTypeSet    = 3
	RDBTypeExpire = 0xFD
	RDBTypeEOF    = 0xFF
)

type RDBFileHeader struct {
	Magic     [5]byte
	Version   uint16
	Timestamp int64
	KeyCount  int32
}

type RDBEntry struct {
	Key      string
	Type     byte
	Value    interface{}
	ExpireAt int64
}

type Serializer struct{}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) SerializeValue(value *storage.StorageValue) ([]byte, error) {
	switch value.Type {
	case storage.StringType:
		return s.serializeString(value.Data.(string))
	case storage.HashType:
		return s.serializeHash(value.Data.(*storage.HashData))
	case storage.ListType:
		return s.serializeList(value.Data.(*storage.ListData))
	case storage.SetType:
		return s.serializeSet(value.Data.(*storage.SetData))
	default:
		return nil, fmt.Errorf("unsupported value type: %v", value.Type)
	}
}

func (s *Serializer) DeserializeValue(valueType storage.ValueType, data []byte) (interface{}, error) {
	switch valueType {
	case storage.StringType:
		return s.deserializeString(data)
	case storage.HashType:
		return s.deserializeHash(data)
	case storage.ListType:
		return s.deserializeList(data)
	case storage.SetType:
		return s.deserializeSet(data)
	default:
		return nil, fmt.Errorf("unsupported value type: %v", valueType)
	}
}

func (s *Serializer) serializeString(str string) ([]byte, error) {
	length := uint32(len(str))
	data := make([]byte, 4+len(str))
	binary.BigEndian.PutUint32(data[0:4], length)
	copy(data[4:], []byte(str))
	return data, nil
}

func (s *Serializer) deserializeString(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("invalid string data length")
	}
	length := binary.BigEndian.Uint32(data[0:4])
	if uint32(len(data)-4) != length {
		return "", fmt.Errorf("string length mismatch")
	}

	return string(data[4 : 4+length]), nil
}

func (s *Serializer) serializeHash(hash *storage.HashData) ([]byte, error) {
	fields := hash.Fields()

	fieldsCount := uint32(len(fields))

	totalSize := 4
	for field, val := range fields {
		totalSize += 4 + len(field) + len(val)
	}

	data := make([]byte, totalSize)
	binary.BigEndian.PutUint32(data[0:4], fieldsCount)
	offset := 4

	for field, val := range fields {
		fieldLen := uint32(len(field))
		binary.BigEndian.PutUint32(data[offset:offset+4], fieldLen)
		offset += 4
		copy(data[offset:offset+len(field)], []byte(field))
		offset += len(field)

		valLen := uint32(len(val))
		binary.BigEndian.PutUint32(data[offset:offset+4], valLen)
		offset += 4
		copy(data[offset:offset+len(val)], []byte(val))
		offset += len(val)
	}

	return data, nil
}

func (s *Serializer) deserializeHash(data []byte) (*storage.HashData, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid hash data length")
	}

	fieldsCount := binary.BigEndian.Uint32(data[0:4])
	hash := storage.NewHashData()
	offset := 4
	for range fieldsCount {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("invalid field length")
		}

		fieldLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		if offset+int(fieldLen) > len(data) {
			return nil, fmt.Errorf("invalid field data")
		}
		field := string(data[offset : offset+int(fieldLen)])
		offset += int(fieldLen)

		if offset+4 > len(data) {
			return nil, fmt.Errorf("invalid value length")
		}

		valueLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		if offset+int(valueLen) > len(data) {
			return nil, fmt.Errorf("invalid value data")
		}
		value := string(data[offset : offset+int(valueLen)])
		offset += int(valueLen)

		hash.Set(field, value)
	}

	return hash, nil
}

func (s *Serializer) serializeList(list *storage.ListData) ([]byte, error) {
	elements := list.GetAll()

	elementsCount := uint32(len(elements))

	totalSize := 4
	for _, element := range elements {
		totalSize += 4 + len(element)
	}

	data := make([]byte, totalSize)
	binary.BigEndian.PutUint32(data[0:4], elementsCount)
	offset := 4

	for _, element := range elements {
		elementLen := uint32(len(element))
		binary.BigEndian.PutUint32(data[offset:offset+4], elementLen)
		offset += 4
		copy(data[offset:offset+len(element)], []byte(element))
		offset += int(elementLen)
	}

	return data, nil
}

func (s *Serializer) deserializeList(data []byte) (*storage.ListData, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid list data length")
	}

	elementsCount := binary.BigEndian.Uint32(data[0:4])
	offset := 4
	list := storage.NewListData()

	for range elementsCount {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("invalid element length")
		}

		elementLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		if offset+int(elementLen) > len(data) {
			return nil, fmt.Errorf("invalid element data")
		}
		element := string(data[offset : offset+int(elementLen)])
		offset += int(elementLen)

		list.PushRight(element)
	}

	return list, nil
}

func (s *Serializer) serializeSet(set *storage.SetData) ([]byte, error) {
	members := set.Members()

	membersCount := uint32(len(members))

	totalSize := 4
	for _, member := range members {
		totalSize += 4 + len(member)
	}

	data := make([]byte, totalSize)
	binary.BigEndian.PutUint32(data[0:4], membersCount)
	offset := 4

	for _, member := range members {
		memberLen := uint32(len(member))
		binary.BigEndian.PutUint32(data[offset:offset+4], memberLen)
		offset += 4
		copy(data[offset:offset+int(memberLen)], []byte(member))
		offset += int(memberLen)
	}

	return data, nil
}

func (s *Serializer) deserializeSet(data []byte) (*storage.SetData, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid set data length")
	}

	membersCount := binary.BigEndian.Uint32(data[0:4])
	offset := 4
	set := storage.NewSetData()

	for range membersCount {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("invalid member length")
		}

		memberLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		if offset+int(memberLen) > len(data) {
			return nil, fmt.Errorf("invalid member data")
		}
		member := string(data[offset : offset+int(memberLen)])
		offset += int(memberLen)

		set.Add(member)
	}

	return set, nil
}

func (s *Serializer) SerializerCommand(command []string) ([]byte, error) {
	return json.Marshal(command)
}

func (s *Serializer) DeserializeCommand(data []byte) ([]string, error) {
	var command []string
	err := json.Unmarshal(data, &command)
	return command, err
}

func (s *Serializer) ConvertToRDB(valueType storage.ValueType) byte {
	switch valueType {
	case storage.StringType:
		return RDBTypeString
	case storage.HashType:
		return RDBTypeHash
	case storage.ListType:
		return RDBTypeList
	case storage.SetType:
		return RDBTypeSet
	default:
		return RDBTypeString
	}
}

func (s *Serializer) ConvertFromRDB(rdbType byte) storage.ValueType {
	switch rdbType {
	case RDBTypeString:
		return storage.StringType
	case RDBTypeHash:
		return storage.HashType
	case RDBTypeList:
		return storage.ListType
	case RDBTypeSet:
		return storage.SetType
	default:
		return storage.StringType
	}
}

func (s *Serializer) EncodeLength(length int) []byte {
	if length < 64 {
		return []byte{byte(length)}
	} else if length < 16384 {
		bytes := make([]byte, 2)
		bytes[0] = byte((length >> 8) | 0x40)
		bytes[1] = byte(length & 0xFF)
		return bytes
	} else {
		bytes := make([]byte, 5)
		bytes[0] = 0x80
		binary.BigEndian.PutUint32(bytes[0:5], uint32(length))
		return bytes
	}
}

func (s *Serializer) DecodeLength(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}

	firstByte := data[0]
	if (firstByte & 0x80) == 0 {
		return int(firstByte & 0x3F), 1
	} else if (firstByte & 0x40) == 0 {
		if len(data) < 2 {
			return 0, 0
		}
		length := int(firstByte&0x3F)<<8 | int(data[1])
		return length, 2
	} else {
		if len(data) < 5 {
			return 0, 0
		}
		length := int(binary.BigEndian.Uint32(data[1:5]))
		return length, 5
	}
}

func (s *Serializer) GetExpireTimestamp(expireAt time.Time) int64 {
	if expireAt.IsZero() {
		return 0
	}
	return expireAt.Unix()
}

func (s *Serializer) ParseExpireTimestamp(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}
