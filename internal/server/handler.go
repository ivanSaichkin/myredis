package server

import (
	"ivanSaichkin/myredis/internal/command"
	"ivanSaichkin/myredis/internal/protocol"
	"ivanSaichkin/myredis/internal/storage"
	"net"
)

type Handler struct {
	storage  storage.Storage
	executor *command.Executor
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{
		storage:  store,
		executor: command.NewExecutor(store),
	}
}

func (h *Handler) HandleConnection(conn net.Conn) error {
	reader := protocol.NewRESPReader(conn)
	writer := protocol.NewRESPWriter(conn)
	parser := command.NewParser(reader)

	for {
		cmd, err := parser.ParseCommand()
		if err != nil {
			if err == protocol.ErrInvalidSyntax {
				respErr := protocol.Value{
					Type: protocol.Error,
					Str:  "ERR Protocol error: invalid syntax",
				}

				if err := writer.Write(respErr); err != nil {
					return err
				}
				writer.Flush()
				continue
			}
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}

		resp := h.executor.Execute(cmd)

		if err := writer.Write(resp); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
	}
}
