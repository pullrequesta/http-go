package response

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

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

// WriteStatusLine builds and writes the status line based on the
// statusCode provided, returns error if any
//
// The [RFC 9112 Section 4] defines status-line as:
//
//	status-line = HTTP-version SP status-code SP [ reason-phrase ]
//	HTTP-version = HTTP-name "/" DIGIT "." DIGIT
//	HTTP-name = %x48.54.54.50 ; HTTP
//	status-code = 3DIGIT
//	reason-phrase = 1*( HTAB / SP / VCHAR / obs-text )
//
// [RFC 9112 Section 4]: https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
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

// WriteHeaders writes [headers.HTTPHeaders] in http format,
// returns error if any.
//
// The http format is as follows:
//
//	field-name: value\r\n
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

// WriteChunkedBody writes the body in chunks,
// returns the number of bytes written and error if any.
//
// see also [WriteChunkedBodyDone]
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	var err error
	trimmed, _ := bytes.CutSuffix(p, []byte("\n"))
	chunkSize := fmt.Sprintf("%x\r\n", len(trimmed))
	_, err = w.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}

	chunkData := fmt.Sprintf("%s\r\n", trimmed)
	var n int
	n, err = w.Write([]byte(chunkData))
	if err != nil {
		return 0, err
	}

	return n, nil

}

// WriteChunkedBodyDone writes the "0\r\n\r\n",
// returns the number of bytes written and error if any.
//
// see also [WriteChunkedBody]
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.Write([]byte("0\r\n\r\n"))

}

// GetDefaultHeaders creates new default headers and
// returns [headers.HTTPHeaders] them based on the
// content length provided
//
// It creates the following headers:
//   - Content-Length: [contentLen]
//   - Connection: close
//   - Content-Type: text/html
func GetDefaultHeaders(contentLen int) headers.HTTPHeaders {
	hdr := headers.NewHeaders()

	contentlength := strconv.Itoa(contentLen)
	hdr.Set("Content-Length", contentlength)
	hdr.Set("Connection", "close")
	hdr.Set("Content-Type", "text/html")
	return hdr
}
