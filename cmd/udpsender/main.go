package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const address = ":42069"

func IsNewLine(r rune) bool {
	return r == '\n'
}

func MUST[T any](x T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return x
}

func main() {

	raddr := MUST(net.ResolveUDPAddr("udp", address))

	fmt.Println("UDP address:", raddr)

	conn := MUST(net.DialUDP("udp", nil, raddr))

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing UDP connection: %v", err)
		}
	}()

	// Create a new buffered reader that reads from standard input
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Printf("➤")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error%s\n", err.Error())
		}
		line = strings.TrimRightFunc(line, IsNewLine)

		n, err := conn.Write([]byte(line))
		if err != nil {
			log.Printf("error writing to UDP connection: %s\n", err)
		}
		fmt.Printf("written %d bytes to the network\n", n)

	}

}
