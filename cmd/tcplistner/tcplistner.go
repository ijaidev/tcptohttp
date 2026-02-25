package tcplistner

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

const inputFilePath = "messages.txt"

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

		linesChan := getLinesChannel(conn)
		for line := range linesChan {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("connection has been closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {

	linesChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(linesChan)
		currentLineContents := ""
		for {
			buffer := make([]byte, 8, 8)
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
