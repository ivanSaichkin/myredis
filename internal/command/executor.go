package command

import (
	"fmt"
	"ivanSaichkin/myredis/internal/protocol"
	"ivanSaichkin/myredis/internal/storage"
	"strconv"
	"strings"
	"time"
)

type Executor struct {
	Storage storage.Storage
}

func NewExecutor(store storage.Storage) *Executor {
	return &Executor{
		Storage: store,
	}
}

func (e *Executor) Execute(cmd *Command) protocol.Value {
	switch cmd.Name {
	case "PING":
		return e.ping(cmd)
	case "SET":
		return e.set(cmd)
	case "GET":
		return e.get(cmd)
	case "DEL":
		return e.del(cmd)
	case "EXISTS":
		return e.exists(cmd)
	case "EXPIRE":
		return e.exists(cmd)
	case "CLEAR", "FLUSHDB":
		return e.clear(cmd)
	default:
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR unknown command '" + cmd.Name + "'",
		}
	}
}

func (e *Executor) ping(cmd *Command) protocol.Value {
	if len(cmd.Args) > 0 {
		return protocol.Value{
			Type: protocol.BulkString,
			Bulk: cmd.Args[0],
		}
	}

	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  "PONG",
	}
}

func (e *Executor) set(cmd *Command) protocol.Value {
	if len(cmd.Args) < 2 {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR wrong number of arguments for 'set' command",
		}
	}

	key := cmd.Args[0]
	value := cmd.Args[1]

	var ttl time.Duration
	for i := 2; i < len(cmd.Args); i++ {
		if strings.ToUpper(cmd.Args[i]) == "EX" && i+1 < len(cmd.Args) {
			seconds, err := strconv.Atoi(cmd.Args[i+1])
			if err != nil {
				return protocol.Value{
					Type: protocol.Error,
					Str:  "ERR invalid expire time",
				}
			}
			ttl = time.Duration(seconds) * time.Second
			break
		}
	}

	var err error
	if ttl > 0 {
		err = e.Storage.SetWithTTL(key, value, ttl)
	} else {
		err = e.Storage.Set(key, value)
	}

	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR" + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  "OK",
	}
}

func (e *Executor) get(cmd *Command) protocol.Value {
	if len(cmd.Args) != 1 {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR wrong number of arguments for 'get' command",
		}
	}

	val, err := e.Storage.Get(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound || err == storage.ErrKeyExpired {
			return protocol.Value{
				Type:   protocol.BulkString,
				IsNull: true,
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR" + err.Error(),
		}
	}

	strVal, ok := val.Value.(string)
	if !ok {
		strVal = fmt.Sprintf("%v", val.Value)
	}

	return protocol.Value{
		Type: protocol.BulkString,
		Bulk: strVal,
	}
}

func (e *Executor) del(cmd *Command) protocol.Value {
	if len(cmd.Args) < 1 {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR wrong number of arguments for 'del' command",
		}
	}

	deletedCount := 0
	for _, key := range cmd.Args {
		if e.Storage.Delete(key) {
			deletedCount++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  deletedCount,
	}
}

func (e *Executor) exists(cmd *Command) protocol.Value {
	if len(cmd.Args) < 1 {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR wrong number of arguments for 'exists' command",
		}
	}

	existsCount := 0
	for _, key := range cmd.Args {
		if e.Storage.Exists(key) {
			existsCount++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  existsCount,
	}
}

func (e *Executor) expire(cmd *Command) protocol.Value {
	if len(cmd.Args) != 2 {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR wrong number of arguments for 'expire' command",
		}
	}

	key := cmd.Args[0]
	seconds, err := strconv.Atoi(cmd.Args[1])
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR invalid expire time",
		}
	}

	success := e.Storage.Expire(key, time.Duration(seconds)*time.Second)
	result := 0
	if success {
		result = 1
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  result,
	}
}

func (e *Executor) clear(cmd *Command) protocol.Value {
	e.Storage.Clear()
	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  "OK",
	}
}
