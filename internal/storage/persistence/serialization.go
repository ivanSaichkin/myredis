package persistence

import (
	"encoding/binary"
	"fmt"
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
