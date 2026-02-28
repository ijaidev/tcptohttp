package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ijaidev/httpfromtcp/internal/request"
	"github.com/ijaidev/httpfromtcp/internal/response"
)

type Server struct {
	listner  net.Listener
	isClosed atomic.Bool
	handler  Handler
}

type HandlerError struct {
	Status      response.StatusCode
	Message     string
	ContentType string
}

type Handler func(w *response.Writer, req *request.Request) error

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listner, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	server := &Server{
		listner: listner,
		handler: handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listner.Close()
}

func (s *Server) listen() {
	for {

		if s.isClosed.Load() {
			break
		}
		conn, err := s.listner.Accept()

		if err != nil {
			panic(err)
		}

		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)

	writer := response.Writer{
		Conn:       conn,
		WriteState: response.StatusLine,
	}

	if err != nil {
		panic(err)
	}

	err = s.handler(&writer, req)

	if err != nil {
		fmt.Println(err)
	}

	conn.Close()
}

func (hr *HandlerError) WriteToResponse(w *response.Writer) error {
	err := w.WriteStatusLine(hr.Status)

	if err != nil {
		return err
	}

	if hr.ContentType == "" {
		hr.ContentType = "text/plain"
	}
	headers := response.GetDefaultHeaders(len(hr.Message))
	headers.Set("Content-Type", hr.ContentType)
	err = w.WriteHeaders(headers)

	if err != nil {
		return err
	}

	err = w.WriteBody([]byte(hr.Message))
	return err
}
