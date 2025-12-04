package main

import (
	"fmt"
	"httpfromtcp/internal"
	"log"
	"net"
)

const ListenAddr string = ":42069"

// Must helper function
func MUST[T any](arg T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return arg
}

func main() {

	ln := MUST(net.Listen("tcp", ListenAddr))

	defer func() {
		if err := ln.Close(); err != nil {
			log.Printf("error closing the TCP listener: %v", err)
		}
	}()

	for {
		conn := MUST(ln.Accept())

		fmt.Println("Accepted the TCP connection from", conn.RemoteAddr())

		msg := MUST(internal.MessageFromReader(conn))
		r, ok := msg.(*internal.Request)
		if !ok {
			log.Println("Did not receive request!")
			continue
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
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")

	}

}
