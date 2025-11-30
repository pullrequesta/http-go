package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

const address = ":42069"
const bufSize = 1024

func Must[T any](x T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return x
}

func main() {

	fmt.Println("starting udp server")
	addr := Must(net.ResolveUDPAddr("udp", address))
	conn := Must(net.ListenUDP("udp", addr))

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing UDP connection: %v", err)
		}
	}()

	buf := make([]byte, bufSize)
	for {
		n, caddr, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}
		if n == 0 {
			continue
		}

		var msg string
		if i := bytes.Index(buf[:n], []byte("\n")); i != -1 {
			msg = string(buf[:i])
		} else {
			msg = string(buf[:n])
		}

		fmt.Printf("[%s]: %s\n", caddr.String(), msg)
		if msg == "close" {
			fmt.Printf("[%s] asked to close\n", caddr.String())
			break
		}
	}
	fmt.Println("UDP server closing")
}
