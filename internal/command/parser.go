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
	val, err := p.reader.Read()
	if err != nil {
		return nil, err
	}

	if val.Type != protocol.Array || val.IsNull {
		return nil, protocol.ErrInvalidSyntax
	}

	if len(val.Array) == 0 {
		return nil, protocol.ErrInvalidSyntax
	}

	cmdName, err := val.Array[0].String()
	if err != nil {
		return nil, err
	}

	args := make([]string, len(val.Array)-1)
	for i := 1; i < len(val.Array); i++ {
		arg, err := val.Array[i].String()
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
