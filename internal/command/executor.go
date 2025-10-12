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
	storage   storage.Storage
	validator *Validator
}

func NewExecutor(store storage.Storage) *Executor {
	return &Executor{
		storage:   store,
		validator: NewValidator(store),
	}
}

func (e *Executor) Execute(cmd *Command) protocol.Value {
	if err := e.validator.ValidateCommand(cmd); err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	switch cmd.Name {
	// String commands
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
		return e.expire(cmd)
	case "TTL":
		return e.ttl(cmd)
	case "TYPE":
		return e.keyType(cmd)

	// Hash commands
	case "HSET":
		return e.hset(cmd)
	case "HGET":
		return e.hget(cmd)
	case "HDEL":
		return e.hdel(cmd)
	case "HEXISTS":
		return e.hexists(cmd)
	case "HGETALL":
		return e.hgetall(cmd)
	case "HKEYS":
		return e.hkeys(cmd)
	case "HLEN":
		return e.hlen(cmd)

	// List commands
	case "LPUSH":
		return e.lpush(cmd)
	case "RPUSH":
		return e.rpush(cmd)
	case "LPOP":
		return e.lpop(cmd)
	case "RPOP":
		return e.rpop(cmd)
	case "LLEN":
		return e.llen(cmd)
	case "LRANGE":
		return e.lrange(cmd)

	// Set commands
	case "SADD":
		return e.sadd(cmd)
	case "SREM":
		return e.srem(cmd)
	case "SISMEMBER":
		return e.sismember(cmd)
	case "SMEMBERS":
		return e.smembers(cmd)
	case "SCARD":
		return e.scard(cmd)
	case "SINTER":
		return e.sinter(cmd)

	// Utility commands
	case "KEYS":
		return e.keys(cmd)
	case "FLUSHDB", "CLEAR":
		return e.clear(cmd)

	default:
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR unknown command '" + cmd.Name + "'",
		}
	}
}

// String commands

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
		err = e.storage.SetWithTTL(key, value, ttl)
	} else {
		err = e.storage.Set(key, value)
	}

	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  "OK",
	}
}

func (e *Executor) get(cmd *Command) protocol.Value {
	val, err := e.storage.Get(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound || err == storage.ErrKeyExpired {
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

	if val.Type != storage.StringType {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "WRONGTYPE Operation against a key holding the wrong kind of value",
		}
	}

	strVal, ok := val.Data.(string)
	if !ok {
		strVal = fmt.Sprintf("%v", val.Data)
	}

	return protocol.Value{
		Type: protocol.BulkString,
		Bulk: strVal,
	}
}

func (e *Executor) del(cmd *Command) protocol.Value {
	deletedCount := 0
	for _, key := range cmd.Args {
		if e.storage.Delete(key) {
			deletedCount++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  deletedCount,
	}
}

func (e *Executor) exists(cmd *Command) protocol.Value {
	existsCount := 0
	for _, key := range cmd.Args {
		if e.storage.Exists(key) {
			existsCount++
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  existsCount,
	}
}

func (e *Executor) expire(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	seconds, err := strconv.Atoi(cmd.Args[1])
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR invalid expire time",
		}
	}

	success := e.storage.Expire(key, time.Duration(seconds)*time.Second)
	result := 0
	if success {
		result = 1
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  result,
	}
}

func (e *Executor) ttl(cmd *Command) protocol.Value {
	ttl, err := e.storage.TTL(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type: protocol.Integer,
				Num:  -2,
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	if ttl == -1 {
		// No TTL
		return protocol.Value{
			Type: protocol.Integer,
			Num:  -1,
		}
	}

	// Convert to seconds
	return protocol.Value{
		Type: protocol.Integer,
		Num:  int(ttl / time.Second),
	}
}

func (e *Executor) keyType(cmd *Command) protocol.Value {
	keyType, err := e.storage.Type(cmd.Args[0])
	if err != nil {
		if err == storage.ErrKeyNotFound {
			return protocol.Value{
				Type: protocol.SimpleString,
				Str:  "none",
			}
		}
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  keyType.String(),
	}
}

// Utility commands

func (e *Executor) keys(cmd *Command) protocol.Value {
	pattern := cmd.Args[0]
	allKeys := e.storage.Keys()

	// Simple pattern matching (only * supported for now)
	var matchedKeys []string
	if pattern == "*" {
		matchedKeys = allKeys
	} else {
		// Basic pattern matching
		for _, key := range allKeys {
			if simpleMatch(key, pattern) {
				matchedKeys = append(matchedKeys, key)
			}
		}
	}

	// Convert to RESP array
	result := make([]protocol.Value, len(matchedKeys))
	for i, key := range matchedKeys {
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

func (e *Executor) clear(cmd *Command) protocol.Value {
	e.storage.Clear()
	return protocol.Value{
		Type: protocol.SimpleString,
		Str:  "OK",
	}
}

// Helper functions

// simpleMatch does basic pattern matching with * wildcard
func simpleMatch(s, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Convert pattern to regex-like but simple
	// Only handles * at the end for now
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(s, prefix)
	}

	return s == pattern
}
