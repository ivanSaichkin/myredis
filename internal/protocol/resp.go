package protocol

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

var (
	ErrInvalidSyntax      = errors.New("resp: invalid syntax")
	ErrUnsupportedType    = errors.New("resp: unsupported type")
	ErrInvalidBulkString  = errors.New("resp: invalid bulk string")
	ErrInvalidArrayLength = errors.New("resp: invalid array length")
)

const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

type Value struct {
	Type   byte
	Str    string
	Num    int
	Bulk   string
	Array  []Value
	IsNull bool
}

type RESPReader struct {
	reader *bufio.Reader
}

type RESPWriter struct {
	writer *bufio.Writer
}

func NewRESPReader(r io.Reader) *RESPReader {
	return &RESPReader{
		reader: bufio.NewReader(r),
	}
}

func NewRESPWriter(w io.Writer) *RESPWriter {
	return &RESPWriter{
		writer: bufio.NewWriter(w),
	}
}

func (r *RESPReader) Read() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	if len(line) == 0 {
		return Value{}, ErrInvalidSyntax
	}

	switch line[0] {
	case SimpleString:
		return Value{Type: SimpleString, Str: string(line[1:])}, nil
	case Error:
		return Value{Type: Error, Str: string(line[1:])}, nil
	case Integer:
		num, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return Value{}, ErrInvalidSyntax
		}
		return Value{Type: Integer, Num: num}, nil
	case BulkString:
		return r.readBulkString(line)
	case Array:
		return r.readArray(line)
	default:
		return Value{}, ErrUnsupportedType
	}
}

func (r *RESPReader) readLine() ([]byte, error) {
	line, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, ErrInvalidSyntax
	}

	return line[:len(line)-2], nil
}

func (r *RESPReader) readBulkString(line []byte) (Value, error) {
	lenght, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return Value{}, ErrInvalidSyntax
	}

	if lenght == -1 {
		return Value{Type: BulkString, IsNull: true}, nil
	}

	if lenght < 0 {
		return Value{}, ErrInvalidBulkString
	}

	data := make([]byte, lenght+2)
	if _, err := io.ReadFull(r.reader, data); err != nil {
		return Value{}, err
	}

	if data[lenght] != '\r' || data[lenght+1] != '\n' {
		return Value{}, ErrInvalidBulkString
	}

	return Value{
		Type:   BulkString,
		Bulk:   string(data[:lenght]),
		IsNull: false,
	}, nil
}

func (r *RESPReader) readArray(line []byte) (Value, error) {
	lenght, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return Value{}, ErrInvalidSyntax
	}

	if lenght == -1 {
		return Value{Type: Array, IsNull: true}, nil
	}

	if lenght < 0 {
		return Value{}, ErrInvalidArrayLength
	}

	data := make([]Value, lenght)
	for i := range lenght {
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		data[i] = val
	}

	return Value{
		Type:   Array,
		Array:  data,
		IsNull: false,
	}, nil
}
