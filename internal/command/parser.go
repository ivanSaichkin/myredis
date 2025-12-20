package command

import (
	"ivanSaichkin/myredis/internal/protocol"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

type Parser struct {
	reader *protocol.RESPReader
}

func NewParser(reader *protocol.RESPReader) *Parser {
	return &Parser{
		reader: reader,
	}
}

func (p *Parser) ParseCommand() (*Command, error) {
	value, err := p.reader.Read()
	if err != nil {
		return nil, err
	}

	if value.Type != protocol.Array || value.IsNull {
		return nil, protocol.ErrInvalidSyntax
	}

	if len(value.Array) == 0 {
		return nil, protocol.ErrInvalidSyntax
	}

	cmdName, err := value.Array[0].String()
	if err != nil {
		return nil, err
	}

	args := make([]string, len(value.Array)-1)
	for i := 1; i < len(value.Array); i++ {
		arg, err := value.Array[i].String()
		if err != nil {
			return nil, err
		}
		args[i-1] = arg
	}

	return &Command{
		Name: strings.ToUpper(cmdName),
		Args: args,
	}, nil
}
