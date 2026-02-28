package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ijaidev/httpfromtcp/internal/headers"
	"github.com/ijaidev/httpfromtcp/internal/request"
	"github.com/ijaidev/httpfromtcp/internal/response"
	"github.com/ijaidev/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer srv.Close()
	fmt.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) error {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range req.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
	fmt.Println("Body:")
	fmt.Println(string(req.Body))

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		return proxyHandler(w, req)
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return handler400(w, req)
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		return handler500(w, req)
	}
	if req.RequestLine.RequestTarget == "/video" {
		return handlerVideo(w, req)
	}
	return handler200(w, req)
}

func proxyHandler(w *response.Writer, req *request.Request) error {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		return handler500(w, req)
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOk)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	h.Remove("Connection")
	w.WriteHeaders(h)

	fullBody := make([]byte, 0)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			fullBody = append(fullBody, buffer[:n]...)
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()

	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	trailers := headers.NewHeaders()
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Override("X-Content-SHA256", sha256)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")

	return nil
}

func handler500(w *response.Writer, _ *request.Request) error {
	resErr := &server.HandlerError{
		Status: response.StatusBad,
		Message: `<html>
  <head>
	<title>400 Bad Request</title>
  </head>
  <body>
	<h1>Bad Request</h1>
	<p>Your request honestly kinda sucked.</p>
  </body>
</html>`,
		ContentType: "text/html",
	}
	return resErr.WriteToResponse(w)
}

func handler400(w *response.Writer, _ *request.Request) error {
	resErr := &server.HandlerError{
		Status: response.StatusServerError,
		Message: `<html>
  <head>
	<title>500 Internal Server Error</title>
  </head>
  <body>
	<h1>Internal Server Error</h1>
	<p>Okay, you know what? This one is on me.</p>
  </body>
</html>`,
		ContentType: "text/html",
	}
	return resErr.WriteToResponse(w)
}

func handler200(w *response.Writer, _ *request.Request) error {
	s := `<html>
  <head>
	<title>200 OK</title>
  </head>
  <body>
	<h1>Success!</h1>
	<p>Your request was an absolute banger.</p>
  </body>
</html>`

	fmt.Println("I'm here")
	err := w.WriteStatusLine(response.StatusOk)
	if err != nil {
		panic(err)
	}

	headers := response.GetDefaultHeaders(len(s))
	headers.Set("Content-Type", "text/html")

	err = w.WriteHeaders(headers)
	if err != nil {
		panic(err)
	}

	return w.WriteBody([]byte(s))
}

func handlerVideo(w *response.Writer, _ *request.Request) error {
	file, err := os.Open("./assets/vim.mp4")
	if err != nil {
		return err
	}

	buffer, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	err = w.WriteStatusLine(response.StatusOk)
	if err != nil {
		panic(err)
	}

	headers := response.GetDefaultHeaders(len(buffer))
	headers.Set("Content-Type", "video/mp4")

	err = w.WriteHeaders(headers)
	if err != nil {
		panic(err)
	}

	return w.WriteBody([]byte(buffer))
}
