package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	adrr, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")

	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, adrr)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		conn.Write([]byte(input))
	}
}
