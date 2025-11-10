package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

const ListenAddr string = ":42069"

func main() {

	ln, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		log.Printf("error listening for TCP traffic: %s\n", err.Error())
	}

	defer func() {
		if err := ln.Close(); err != nil {
			log.Printf("error closing the TCP listener: %v", err)
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error connecting to the TCP listener: %s\n", err.Error())
		}

		fmt.Println("Accepted the TCP connection from", conn.RemoteAddr())

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("error: %s\n", err.Error())
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range r.Headers.Headers {
			fmt.Println("-", key, ":", val)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")

	}

}

// const bufferSize = 8
// const address = ":42069"

// func getLinesChannel(conn io.ReadCloser) <-chan string {
// 	data := make([]byte, bufferSize)
// 	line := ""
// 	ch := make(chan string)

// 	go func() {
// 		defer close(ch)
// 		for {
// 			n, err := conn.Read(data)
// 			if err == io.EOF {
// 				break
// 			}
// 			data := data[:n]
// 			if idx := bytes.Index(data, []byte("\n\r")); idx != -1 {
// 				line += string(data[:idx])
// 				ch <- line
// 				line = ""
// 				data = data[idx+2:]

// 			}

// 			line += string(data)

// 		}
// 		if err := conn.Close(); err != nil {
// 			fmt.Println(err.Error())
// 		}
// 		ch <- line
// 	}()

// 	return ch
// }

// func main() {

// 	ln, err := net.Listen("tcp", address)
// 	if err != nil {
// 		log.Printf("error listening for TCP traffic: %s\n", err.Error())
// 	}

// 	defer func() {
// 		if err := ln.Close(); err != nil {
// 			log.Printf("error closing TCP listener: %v", err)
// 		}
// 	}()

// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Printf("error: %s\n", err.Error())
// 		}

// 		fmt.Println("Accepted TCP connection from", conn.RemoteAddr())

// 		for line := range getLinesChannel(conn) {
// 			fmt.Println(line)
// 		}
// 		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
// 	}

// }
