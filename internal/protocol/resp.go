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

// Read function
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

// Write function
func (w *RESPWriter) Write(value Value) error {
	switch value.Type {
	case SimpleString:
		return w.writeSimpleString(value.Str)
	case Error:
		return w.writeError(value.Str)
	case Integer:
		return w.writeInteger(value.Num)
	case BulkString:
		return w.writeBulkString(value.Bulk)
	case Array:
		return w.writeArray(value.Array)
	default:
		return ErrUnsupportedType
	}
}

func (w *RESPWriter) writeSimpleString(s string) error {
	if _, err := w.writer.WriteString("+" + s + "\r\n"); err != nil {
		return err
	}

	return nil
}

func (w *RESPWriter) writeError(s string) error {
	if _, err := w.writer.WriteString("-" + s + "\r\n"); err != nil {
		return err
	}

	return nil
}

func (w *RESPWriter) writeInteger(n int) error {
	if _, err := w.writer.WriteString(":" + strconv.Itoa(n) + "\r\n"); err != nil {
		return err
	}

	return nil
}

func (w *RESPWriter) writeBulkString(s string) error {
	if s == "" {
		if _, err := w.writer.WriteString("$-1\r\n"); err != nil {
			return err
		}
	} else {
		if _, err := w.writer.WriteString("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n"); err != nil {
			return err
		}
	}

	return nil
}

func (w *RESPWriter) writeArray(arr []Value) error {
	if _, err := w.writer.WriteString("*" + strconv.Itoa(len(arr)) + "\r\n"); err != nil {
		return err
	}

	for _, item := range arr {
		if err := w.Write(item); err != nil {
			return err
		}
	}

	return nil
}

func (w *RESPWriter) Flush() error {
	return w.writer.Flush()
}

// functions for common resp
func (w *RESPWriter) WriteOK() error {
	return w.writeSimpleString("OK")
}

func (w *RESPWriter) WriteString(s string) error {
	return w.writeBulkString(s)
}

func (w *RESPWriter) WriteInteger(n int) error {
	return w.writeInteger(n)
}

func (w *RESPWriter) WriteNull() error {
	return w.writeBulkString("")
}

func (w *RESPWriter) WriteError(err error) error {
	return w.writeError("ERR" + err.Error())
}
