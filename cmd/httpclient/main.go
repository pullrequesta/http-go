package main

import (
	"flag"
	"fmt"
	"httpfromtcp/internal"

	"log"
	"net"
	"strings"
)

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

	r := internal.NewRequest("GET", "/yourproblem")
	r.SetBody([]byte("Welcome"), "plain/text")
	n, err := connection.Write([]byte(r.String()))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written %d to %s\n", n, connection.RemoteAddr().String())
	msg, err := internal.MessageFromReader(connection)
	if err != nil {
		log.Printf("error parsing response: %v", err)
		return
	}
	resp, ok := msg.(*internal.Response)
	if !ok {
		log.Println("error converting http message to response")
		return
	}
	fmt.Println("Server returned response")
	fmt.Printf("- Status Code: %d\n", resp.ResponseLine.StatusCode)
	fmt.Printf("- Reason: %s\n", resp.ResponseLine.ReasonPhrase)
	for k, v := range resp.Headers.HeadersMap {
		fmt.Printf("- %s: %s\n", k, v)
	}
	fmt.Printf("- Body: %s\n", string(resp.GetBody()))

}
