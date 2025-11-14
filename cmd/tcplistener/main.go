package main

import (
	"fmt"
	"httpfromtcp/internal/request"
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
	// if err != nil {
	// 	log.Printf("error listening for TCP traffic: %s\n", err.Error())
	// }

	defer func() {
		if err := ln.Close(); err != nil {
			log.Printf("error closing the TCP listener: %v", err)
		}
	}()

	for {
		conn := MUST(ln.Accept())
		// if err != nil {
		// 	log.Printf("error connecting to the TCP listener: %s\n", err.Error())
		// }

		fmt.Println("Accepted the TCP connection from", conn.RemoteAddr())

		r := MUST(request.RequestFromReader(conn))
		// if err != nil {
		// 	log.Printf("error: %s\n", err.Error())
		// }
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range r.Headers.Headers {
			fmt.Println("-", key, ":", val)
		}
		fmt.Printf("- Body: %s\n", string(r.Body))

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")

	}

}
