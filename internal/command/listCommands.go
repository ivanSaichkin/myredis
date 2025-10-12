package command

import (
	"ivanSaichkin/myredis/internal/protocol"
	"ivanSaichkin/myredis/internal/storage"
	"strconv"
)

// List commands

func (e *Executor) lpush(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	values := cmd.Args[1:]

	length, err := e.storage.LPush(key, values...)
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  length,
	}
}

func (e *Executor) rpush(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	values := cmd.Args[1:]

	length, err := e.storage.RPush(key, values...)
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  length,
	}
}

func (e *Executor) lpop(cmd *Command) protocol.Value {
	value, err := e.storage.LPop(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type:   protocol.BulkString,
				IsNull: true,
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.BulkString,
		Bulk: value,
	}
}

func (e *Executor) rpop(cmd *Command) protocol.Value {
	value, err := e.storage.RPop(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type:   protocol.BulkString,
				IsNull: true,
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.BulkString,
		Bulk: value,
	}
}

func (e *Executor) llen(cmd *Command) protocol.Value {
	length, err := e.storage.LLen(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type: protocol.Integer,
				Num:  0,
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  length,
	}
}

func (e *Executor) lrange(cmd *Command) protocol.Value {
	start, err := strconv.Atoi(cmd.Args[1])
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR invalid start index",
		}
	}

	stop, err := strconv.Atoi(cmd.Args[2])
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR invalid stop index",
		}
	}

	elements, err := e.storage.LRange(cmd.Args[0], start, stop)
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type:  protocol.Array,
				Array: []protocol.Value{},
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	result := make([]protocol.Value, len(elements))
	for i, element := range elements {
		result[i] = protocol.Value{
			Type: protocol.BulkString,
			Bulk: element,
		}
	}

	return protocol.Value{
		Type:  protocol.Array,
		Array: result,
	}
}
