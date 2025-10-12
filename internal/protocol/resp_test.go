package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestRESPReader_ReadSimpleString(t *testing.T) {
	input := "+OK\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != SimpleString {
		t.Errorf("Expected type SimpleString, got %c", value.Type)
	}

	if value.Str != "OK" {
		t.Errorf("Expected 'OK', got '%s'", value.Str)
	}
}

func TestRESPReader_ReadError(t *testing.T) {
	input := "-ERR something went wrong\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != Error {
		t.Errorf("Expected type Error, got %c", value.Type)
	}

	if value.Str != "ERR something went wrong" {
		t.Errorf("Expected 'ERR something went wrong', got '%s'", value.Str)
	}
}

func TestRESPReader_ReadInteger(t *testing.T) {
	input := ":42\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != Integer {
		t.Errorf("Expected type Integer, got %c", value.Type)
	}

	if value.Num != 42 {
		t.Errorf("Expected 42, got %d", value.Num)
	}
}

func TestRESPReader_ReadBulkString(t *testing.T) {
	input := "$5\r\nhello\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != BulkString {
		t.Errorf("Expected type BulkString, got %c", value.Type)
	}

	if value.Bulk != "hello" {
		t.Errorf("Expected 'hello', got '%s'", value.Bulk)
	}
}

func TestRESPReader_ReadNullBulkString(t *testing.T) {
	input := "$-1\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != BulkString {
		t.Errorf("Expected type BulkString, got %c", value.Type)
	}

	if !value.IsNull {
		t.Error("Expected IsNull to be true")
	}
}

func TestRESPReader_ReadArray(t *testing.T) {
	input := "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
	reader := NewRESPReader(strings.NewReader(input))

	value, err := reader.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value.Type != Array {
		t.Errorf("Expected type Array, got %c", value.Type)
	}

	if len(value.Array) != 2 {
		t.Fatalf("Expected array length 2, got %d", len(value.Array))
	}

	if value.Array[0].Bulk != "hello" {
		t.Errorf("Expected first element 'hello', got '%s'", value.Array[0].Bulk)
	}

	if value.Array[1].Bulk != "world" {
		t.Errorf("Expected second element 'world', got '%s'", value.Array[1].Bulk)
	}
}

func TestRESPWriter_WriteSimpleString(t *testing.T) {
	var buf bytes.Buffer
	writer := NewRESPWriter(&buf)

	err := writer.Write(Value{Type: SimpleString, Str: "OK"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	writer.Flush()

	expected := "+OK\r\n"
	if buf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.String())
	}
}

func TestRESPWriter_WriteBulkString(t *testing.T) {
	var buf bytes.Buffer
	writer := NewRESPWriter(&buf)

	err := writer.Write(Value{Type: BulkString, Bulk: "hello"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	writer.Flush()

	expected := "$5\r\nhello\r\n"
	if buf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.String())
	}
}

func TestRESPWriter_WriteArray(t *testing.T) {
	var buf bytes.Buffer
	writer := NewRESPWriter(&buf)

	array := []Value{
		{Type: BulkString, Bulk: "hello"},
		{Type: BulkString, Bulk: "world"},
	}

	err := writer.Write(Value{Type: Array, Array: array})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	writer.Flush()

	expected := "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
	if buf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.String())
	}
}
