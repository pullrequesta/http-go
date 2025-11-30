package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
)

type Handler func(w *response.Writer, r *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (e HandlerError) Error() string {
	return fmt.Sprintf("%d %s\n", e.StatusCode, e.Message)
}

func Must[T any](x T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return x
}

type TransportProtocol string

const (
	PROTO_UDP TransportProtocol = "udp"
	PROTO_TCP TransportProtocol = "tcp"
)

type ServerOptions struct {
	proto TransportProtocol
	addr  string
}

type Server struct {
	opts     *ServerOptions
	listener net.Listener
	handler  func(w *response.Writer, r *request.Request)
	doneCh   chan bool
}

func defaultServerOptions() *ServerOptions {
	return &ServerOptions{
		proto: PROTO_TCP,
		addr:  ":42069",
	}
}

type ServerOption func(*ServerOptions)

func WithTCP() ServerOption {
	return func(opts *ServerOptions) {
		opts.proto = PROTO_TCP
	}
}

func WithUDP() ServerOption {
	return func(opts *ServerOptions) {
		opts.proto = PROTO_UDP
	}
}

func WithAddr(addr string) ServerOption {
	return func(opts *ServerOptions) {
		opts.addr = addr
	}
}

// NewServer creates a new Server with options provided.
// If no options are provided proto "tcp" will be used
// and default address ":42069" will be used.
func NewServer(opts ...ServerOption) *Server {
	o := defaultServerOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &Server{
		opts:     o,
		listener: nil,
		handler:  nil,
		doneCh:   make(chan bool),
	}
}

func (s *Server) Address() string {
	return s.opts.addr
}

func (s *Server) Serve(hf Handler) error {
	s.handler = hf
	switch s.opts.proto {
	case PROTO_TCP:
		fmt.Println("starting tcp server")
		var err error
		s.listener, err = net.Listen(string(s.opts.proto), s.opts.addr)
		if err != nil {
			return err
		}
		go s.TCPlisten()
	case PROTO_UDP:
		s.listener = nil
		go s.UDPlisten()
	}
	return nil
}

func (s *Server) TCPlisten() {

	for {
		conn := Must(s.listener.Accept())
		fmt.Printf("accepted the TCP connection from: %d", conn.RemoteAddr())

		go s.handleConn(conn)
	}
}

func (s *Server) UDPlisten() {

	fmt.Println("starting udp server")
	addr, err := net.ResolveUDPAddr(string(s.opts.proto), s.opts.addr)
	if err != nil {
		log.Println(err)
		return
	}
	conn, err := net.ListenUDP(string(s.opts.proto), addr)
	if err != nil {
		log.Println(err)
		return
	}
	go s.handleConn(conn)
}

func (s *Server) handleConn(rwc io.ReadWriteCloser) {

	log.Println("hanlding connection")
	defer func() {
		if err := rwc.Close(); err != nil {
			log.Printf("error closing the connection: %v", err)
		}
	}()

	// parse the request from the connection.
	r, err := request.RequestFromReader(rwc)
	if err != nil {
		log.Printf("error parsing request: %v", err)
	}
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

	s.handler(responseWriter, r)

}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		log.Printf("error closing the listener: %v", err)
	}
	return nil

}

func (s *Server) Done() {
	fmt.Printf("Sending True to listener to signal end of life\n")
	s.doneCh <- true
	fmt.Printf("already sent\n")
}
