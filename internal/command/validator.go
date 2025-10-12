package command

import (
	"errors"
	"ivanSaichkin/myredis/internal/storage"
	"strconv"
)

var (
	ErrWrongNumberOfArguments = errors.New("wrong number of arguments")
	ErrInvalidInteger         = errors.New("invalid integer")
	ErrSyntaxError            = errors.New("syntax error")
)

type Validator struct {
	storage storage.Storage
}

func NewValidator(store storage.Storage) *Validator {
	return &Validator{
		storage: store,
	}
}

func (v *Validator) ValidateCommand(cmd *Command) error {
	switch cmd.Name {
	// String commands
	case "SET":
		return v.validateSet(cmd)
	case "GET":
		return v.validateGet(cmd)
	case "DEL":
		return v.validateDel(cmd)
	case "EXISTS":
		return v.validateExists(cmd)
	case "EXPIRE":
		return v.validateExpire(cmd)
	case "TTL":
		return v.validateTTL(cmd)

	// Hash commands
	case "HSET":
		return v.validateHSet(cmd)
	case "HGET":
		return v.validateHGet(cmd)
	case "HDEL":
		return v.validateHDel(cmd)
	case "HEXISTS":
		return v.validateHExists(cmd)
	case "HGETALL":
		return v.validateHGetAll(cmd)
	case "HKEYS":
		return v.validateHKeys(cmd)
	case "HLEN":
		return v.validateHLen(cmd)

	// List commands
	case "LPUSH":
		return v.validateLPush(cmd)
	case "RPUSH":
		return v.validateRPush(cmd)
	case "LPOP":
		return v.validateLPop(cmd)
	case "RPOP":
		return v.validateRPop(cmd)
	case "LLEN":
		return v.validateLLen(cmd)
	case "LRANGE":
		return v.validateLRange(cmd)

	// Set commands
	case "SADD":
		return v.validateSAdd(cmd)
	case "SREM":
		return v.validateSRem(cmd)
	case "SISMEMBER":
		return v.validateSIsMember(cmd)
	case "SMEMBERS":
		return v.validateSMembers(cmd)
	case "SCARD":
		return v.validateSCard(cmd)
	case "SINTER":
		return v.validateSInter(cmd)

	// Utility commands
	case "KEYS":
		return v.validateKeys(cmd)
	case "FLUSHDB", "CLEAR":
		return v.validateClear(cmd)
	case "PING":
		return v.validatePing(cmd)
	case "TYPE":
		return v.validateType(cmd)

	default:
		return nil
	}
}

// String commands validation

func (v *Validator) validateSet(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateGet(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateDel(cmd *Command) error {
	if len(cmd.Args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateExists(cmd *Command) error {
	if len(cmd.Args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateExpire(cmd *Command) error {
	if len(cmd.Args) != 2 {
		return ErrWrongNumberOfArguments
	}

	if _, err := strconv.Atoi(cmd.Args[1]); err != nil {
		return ErrInvalidInteger
	}
	return nil
}

func (v *Validator) validateTTL(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

// Hash commands validation

func (v *Validator) validateHSet(cmd *Command) error {
	if len(cmd.Args) < 3 || len(cmd.Args)%2 != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHGet(cmd *Command) error {
	if len(cmd.Args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHDel(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHExists(cmd *Command) error {
	if len(cmd.Args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHGetAll(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHKeys(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateHLen(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

// List commands validation

func (v *Validator) validateLPush(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateRPush(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateLPop(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateRPop(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateLLen(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateLRange(cmd *Command) error {
	if len(cmd.Args) != 3 {
		return ErrWrongNumberOfArguments
	}

	if _, err := strconv.Atoi(cmd.Args[1]); err != nil {
		return ErrInvalidInteger
	}

	if _, err := strconv.Atoi(cmd.Args[2]); err != nil {
		return ErrInvalidInteger
	}
	return nil
}

// Set commands validation

func (v *Validator) validateSAdd(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateSRem(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateSIsMember(cmd *Command) error {
	if len(cmd.Args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateSMembers(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateSCard(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateSInter(cmd *Command) error {
	if len(cmd.Args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

// Utility commands validation

func (v *Validator) validateKeys(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateClear(cmd *Command) error {
	if len(cmd.Args) != 0 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validatePing(cmd *Command) error {
	if len(cmd.Args) > 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (v *Validator) validateType(cmd *Command) error {
	if len(cmd.Args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}
