package response

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/ijaidev/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOk          StatusCode = 200
	StatusBad         StatusCode = 400
	StatusServerError StatusCode = 500
)

type WriteState int

const (
	StatusLine WriteState = iota
	Headers
	Body
	Trailers
)

type Writer struct {
	WriteState WriteState
	Conn       net.Conn
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {

	if w.WriteState != StatusLine {
		return errors.New("Status line is alreay written")
	}

	line := ""
	switch statusCode {
	case StatusOk:
		{
			line = "HTTP/1.1 200 OK"
		}
	case StatusBad:
		{
			line = "HTTP/1.1 400 Bad Request"
		}
	case StatusServerError:
		{
			line = "HTTP/1.1 500 Internal Server Error"
		}
	default:
		{
			line = fmt.Sprintf("HTTP/1.1 %d ", statusCode)
		}
	}
	fullLine := line + "\r\n"
	_, err := w.Conn.Write([]byte(fullLine))
	if err != nil {
		return err
	}
	w.WriteState = Headers
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {

	if w.WriteState != Headers {
		return errors.New("Headers are alreay written or Being written before status line")
	}

	for key, value := range headers {
		h := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Conn.Write([]byte(h))
		if err != nil {
			return err
		}
	}
	endingLine := "\r\n"
	_, err := w.Conn.Write([]byte(endingLine))
	if err != nil {
		return err
	}
	w.WriteState = Body
	return nil
}

func (w *Writer) WriteBody(p []byte) error {
	if w.WriteState != Body {
		return errors.New("Body should be written after status line and headers")
	}
	_, err := w.Conn.Write(p)
	return err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriteState != Body {
		return 0, fmt.Errorf("cannot write body in state %d", w.WriteState)
	}
	chunkSize := len(p)

	nTotal := 0
	n, err := fmt.Fprintf(w.Conn, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.Conn.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.Conn.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.WriteState != Body {
		return 0, fmt.Errorf("cannot write body in state %d", w.WriteState)
	}
	n, err := w.Conn.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.WriteState = Trailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.WriteState != Trailers {
		return fmt.Errorf("cannot write trailers in state %d", w.WriteState)
	}
	for k, v := range h {
		_, err := fmt.Fprintf(w.Conn, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.Conn.Write([]byte("\r\n"))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.Headers{}

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "Close")
	headers.Set("Content-Type", "text/plain")

	return headers
}
