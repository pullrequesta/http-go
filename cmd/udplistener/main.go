package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

const address = ":42069"
const bufSize = 1024

func MUST[T any](x T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return x
}

func main() {
	fmt.Println("starting udp server")
	addr := MUST(net.ResolveUDPAddr("udp", address))
	conn := MUST(net.ListenUDP("udp", addr))

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
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
