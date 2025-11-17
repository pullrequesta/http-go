package response

import (
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

// Define an enum-like type for status code
type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	Writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {

	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "HTTP/1.1 200 OK"

	case StatusBadRequest:
		reasonPhrase = "HTTP/1.1 400 Bad Request"

	case StatusInternalServerError:
		reasonPhrase = "HTTP/1.1 500 Internal Server Error"
	default:
		reasonPhrase = "HTTP/1.1 500 Internal Server Error"
	}

	_, err := w.Write([]byte(reasonPhrase + "\r\n"))
	return err

}
func GetDefaultHeaders(contentLen int) headers.HTTPHeaders {
	hdr := headers.NewHeaders()

	contentlength := strconv.Itoa(contentLen)
	hdr.Set("Content-Length", contentlength)

	hdr.Set("Connection", "close")
	hdr.Set("Content-Type", "text/html")

	return hdr
}

func (w *Writer) WriteHeaders(headers headers.HTTPHeaders) error {
	hdrs := []byte{}

	for k, v := range headers.HeadersMap {
		sngleHdr := k + ": " + v + "\r\n"
		hdrs = append(hdrs, []byte(sngleHdr)...)
	}
	hdrs = append(hdrs, []byte("\r\n")...)
	_, err := w.Write(hdrs)
	return err

}
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.Write(p)

	return n, err
}
