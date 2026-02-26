package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/ijaidev/httpfromtcp/internal/request"
)

func main() {
	listner, err := net.Listen("tcp", "127.0.0.1:42069")

	if err != nil {
		panic(err)
	}
	defer listner.Close()

	for {
		conn, err := listner.Accept()

		if err != nil {
			panic(err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			panic(err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {

	linesChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(linesChan)
		currentLineContents := ""
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if currentLineContents != "" {
					linesChan <- currentLineContents
					currentLineContents = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			currentLineContents += string(buffer[:n])
			parts := strings.Split(currentLineContents, "\n")

			if len(parts) < 2 {
				continue
			}

			for i := 0; i < len(parts)-1; i++ {
				linesChan <- parts[i]
			}

			currentLineContents = parts[len(parts)-1]
		}
	}()

	return linesChan

}
