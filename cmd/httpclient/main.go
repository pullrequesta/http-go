package main

import (
	"flag"
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
	"strings"
)

func IsNewLine(r rune) bool {
	return r == '\n'
}

func Must[T any](x T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return x
}

func main() {
	addr := flag.String("addr", ":8000", "server address")
	proto := flag.String("proto", "tcp", "protocol to use")
	flag.Parse()

	var connection net.Conn
	switch strings.ToLower(*proto) {

	case "tcp":
		connection = Must(net.Dial("tcp", *addr))

	case "udp":
		raddr := Must(net.ResolveUDPAddr("udp", *addr))
		fmt.Println("UDP address:", raddr)
		connection = Must(net.DialUDP("udp", nil, raddr))
	default:
		log.Fatal("Invalid protocol used. see help")
	}
	defer func() {
		if err := connection.Close(); err != nil {
			log.Printf("error closing UDP connection: %v", err)
		}
	}()

	r := request.NewRequest("GET", "/yourproblem")
	r.SetBody([]byte("Welcome"), "plain/text")
	n, err := connection.Write([]byte(r.String()))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written %d to %s\n", n, connection.RemoteAddr().String())
	buf := make([]byte, 8024)
	for {
		n, err = connection.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n == 0 {
			continue
		}
		log.Printf("Server sent back response %v", string(buf[:n]))
	}
}
