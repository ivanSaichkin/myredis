package server

import (
	"fmt"
	"log"
	"net"
)

type TCPServer struct {
	address string
	handler *Handler
}

func NewTCPServer(addr string, handler *Handler) *TCPServer {
	return &TCPServer{
		address: addr,
		handler: handler,
	}
}

func (t *TCPServer) Start() error {
	listener, err := net.Listen("tcp", t.address)

	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	defer listener.Close()

	log.Printf("Server started on %s", t.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go t.handleConnection(conn)
	}
}

func (t *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Client connected: %s", conn.RemoteAddr())

	if err := t.handler.HandleConnection(conn); err != nil {
		log.Printf("Error handling connection from %s: %v", conn.RemoteAddr(), err)
	}

	log.Printf("Client disconnected: %s", conn.RemoteAddr())
}
