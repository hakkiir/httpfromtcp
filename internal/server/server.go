package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	isRunning atomic.Bool
	listener  net.Listener
}

func Serve(port int) (*Server, error) {

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
	}

	server := Server{
		listener: tcpListener,
	}
	server.isRunning.Store(true)

	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}
	s.isRunning.Store(false)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.isRunning.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}

}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n" + // Status line
		"Content-Type: text/plain\r\n" + // Example header
		"\r\n" + // Blank line to separate headers from the body
		"Hello World!\n" // Body
	conn.Write([]byte(response))
	return
}
