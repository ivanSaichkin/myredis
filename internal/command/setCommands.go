package command

import (
	"ivanSaichkin/myredis/internal/protocol"
	"ivanSaichkin/myredis/internal/storage"
)

// Set commands

func (e *Executor) sadd(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	members := cmd.Args[1:]

	added, err := e.storage.SAdd(key, members...)
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  added,
	}
}

func (e *Executor) srem(cmd *Command) protocol.Value {
	key := cmd.Args[0]
	members := cmd.Args[1:]

	removed, err := e.storage.SRem(key, members...)
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
		Num:  removed,
	}
}

func (e *Executor) sismember(cmd *Command) protocol.Value {
	isMember, err := e.storage.SIsMember(cmd.Args[0], cmd.Args[1])
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
	if isMember {
		result = 1
	}

	return protocol.Value{
		Type: protocol.Integer,
		Num:  result,
	}
}

func (e *Executor) smembers(cmd *Command) protocol.Value {
	members, err := e.storage.SMembers(cmd.Args[0])
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

	result := make([]protocol.Value, len(members))
	for i, member := range members {
		result[i] = protocol.Value{
			Type: protocol.BulkString,
			Bulk: member,
		}
	}

	return protocol.Value{
		Type:  protocol.Array,
		Array: result,
	}
}

func (e *Executor) scard(cmd *Command) protocol.Value {
	count, err := e.storage.SCard(cmd.Args[0])
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
		Num:  count,
	}
}

func (e *Executor) sinter(cmd *Command) protocol.Value {
	members, err := e.storage.SInter(cmd.Args...)
	if err != nil {
		return protocol.Value{
			Type: protocol.Error,
			Str:  "ERR " + err.Error(),
		}
	}

	result := make([]protocol.Value, len(members))
	for i, member := range members {
		result[i] = protocol.Value{
			Type: protocol.BulkString,
			Bulk: member,
		}
	}

	return protocol.Value{
		Type:  protocol.Array,
		Array: result,
	}
}
