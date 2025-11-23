package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
)

type ServerState int

type Server struct {
	listener net.Listener
	handler  func(w *response.Writer, r *request.Request)
	State    ServerState
	doneCh   chan bool
}

type Handler func(w *response.Writer, r *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (e HandlerError) Error() string {
	return fmt.Sprintf("%d %s\n", e.StatusCode, e.Message)
}

func Serve(port int, hf Handler) (*Server, error) {

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: ln,
		handler:  hf,
		doneCh:   make(chan bool),
	}

	// start accepting connections in a goroutine
	go server.listen()

	return server, nil
}

func (s *Server) listen() {

	for {
		// select {
		// case x := <-s.doneCh:
		// 	log.Println("Closing from listener", x)
		// 	return
		// default:
		// Wait for a connection.
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("error connecting to the TCP listener: %v\n", err)
		}

		fmt.Printf("accepted the TCP connection from: %d", conn.RemoteAddr())

		// handle each connection concurrently
		go s.handleConn(conn)
		// }
	}
}

func (s *Server) Done() {
	fmt.Printf("Sending True to listener to signal end of life\n")
	s.doneCh <- true
	fmt.Printf("already sent\n")
}

func (s *Server) handleConn(rwc io.ReadWriteCloser) {

	defer func() {
		if err := rwc.Close(); err != nil {
			log.Printf("error closing the connection: %v", err)
		}
	}()

	// parse the request from the connection
	r, err := request.RequestFromReader(rwc)
	if err != nil {
		log.Printf("error parsing request: %v", err)
	}
	// print the request on the server
	fmt.Printf("Request line:\n")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
	fmt.Printf("Headers:\n")
	for key, val := range r.Headers.HeadersMap {
		fmt.Println("-", key, ":", val)
	}
	fmt.Printf("- Body: %s\n", string(r.Body))

	responseWriter := response.NewWriter(rwc)

	// n, err := responseWriter.WriteChunkedBody([]byte("Welcome"))
	// if err != nil {
	// 	fmt.Printf("%v", err)
	// }
	// fmt.Println(n)

	//create a handler function to write the response
	s.handler(responseWriter, r)

}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		log.Printf("error closing the listener: %v", err)

	}
	return nil
}
