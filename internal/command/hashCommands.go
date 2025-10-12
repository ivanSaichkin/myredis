package command

import (
	"ivanSaichkin/myredis/internal/protocol"
	"ivanSaichkin/myredis/internal/storage"
)

// Hash commands

func (e *Executor) hset(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	fieldsAndValues := cmd.Args[1:]

	totalFields := 0
	for i := 0; i < len(fieldsAndValues); i += 2 {
		if i+1 >= len(fieldsAndValues) {
			break
		}
		field := fieldsAndValues[i]
		value := fieldsAndValues[i+1]

		created, err := e.storage.HSet(key, field, value)
		if err != nil {
			return protocol.Value{
				Type: protocol.Error,
				Str:  "ERR " + err.Error(),
			}
		}
		if created {
			totalFields++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  totalFields,
	}
}

func (e *Executor) hget(cmd *Command) protocol.Value {
	value, err := e.storage.HGet(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if err == storage.ErrKeyNotFound || err == storage.ErrFieldNotFound {
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

func (e *Executor) hdel(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	fields := cmd.Args[1:]

	deletedCount := 0
	for _, field := range fields {
		deleted, err := e.storage.HDel(key, field)
		if err != nil {
			if err == storage.ErrKeyNotFound {
				continue
			}
			return protocol.Value{
				Type: protocol.Error,
				Str:  "ERR " + err.Error(),
			}
		}
		if deleted {
			deletedCount++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  deletedCount,
	}
}

func (e *Executor) hexists(cmd *Command) protocol.Value {
	exists, err := e.storage.HExists(cmd.Args[0], cmd.Args[1])
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

	result := 0
	if exists {
		result = 1
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  result,
	}
}

func (e *Executor) hgetall(cmd *Command) protocol.Value {
	fields, err := e.storage.HGetAll(cmd.Args[0])
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

	result := make([]protocol.Value, 0, len(fields)*2)
	for field, value := range fields {
		result = append(result, protocol.Value{
			Type: protocol.BulkString,
			Bulk: field,
		})
		result = append(result, protocol.Value{
			Type: protocol.BulkString,
			Bulk: value,
		})
	}

	return protocol.Value{
		Type:  protocol.Array,
		Array: result,
	}
}

func (e *Executor) hkeys(cmd *Command) protocol.Value {
	keys, err := e.storage.HKeys(cmd.Args[0])
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

	result := make([]protocol.Value, len(keys))
	for i, key := range keys {
		result[i] = protocol.Value{
			Type: protocol.BulkString,
			Bulk: key,
		}
	}

	return protocol.Value{
		Type:  protocol.Array,
		Array: result,
	}
}

func (e *Executor) hlen(cmd *Command) protocol.Value {
	length, err := e.storage.HLen(cmd.Args[0])
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
