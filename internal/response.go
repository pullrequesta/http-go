package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type HTTPStatusCode int

const (
	StatusOK                  HTTPStatusCode = 200
	StatusBadRequest          HTTPStatusCode = 400
	StatusInternalServerError HTTPStatusCode = 500
)

type Response struct {
	ResponseLine  ResponseLine
	Headers       HTTPHeaders
	ContentLength int
	Body          []byte
}
type ResponseLine struct {
	HTTPVersion  string
	StatusCode   HTTPStatusCode
	ReasonPhrase string
}

func newResponseLine(statusCode HTTPStatusCode, reasonPhrase string) *ResponseLine {
	return &ResponseLine{
		HTTPVersion:  HTTP_VERSION,
		StatusCode:   statusCode,
		ReasonPhrase: reasonPhrase,
	}
}
func NewResponse(statusCode HTTPStatusCode, reasonPhrase string) *Response {
	return &Response{
		ResponseLine: *newResponseLine(statusCode, reasonPhrase),
	}
}

// ParseHeaders parses data in the ParseHeaders function as an input,
// returns number of bytes parsed, bool indicating
// parsing completion and error if any.
func (r *Response) ParseHeaders(data []byte) (int, bool, error) {
	return r.Headers.Parse(data)
}

// GetHeader returns the value of a particular header by its name.
func (r *Response) GetHeader(name string) string {
	return r.Headers.Get(name)
}

// GetContentLength returns the [Content-Length] in the form of an integer.
func (r *Response) GetContentLength() int {
	return r.ContentLength
}

// SetContentLength sets the [Content-Length] field of the header.
func (r *Response) SetContentLength(cl int) {
	r.ContentLength = cl
}

// GetBody returns the body of the response.
func (r *Response) GetBody() []byte {
	return r.Body
}

// SetBody sets the body of the response,
// sets the [Content-Length] according the length of body,
// and sets the [Content-Type] to its values.
func (r *Response) SetBody(body []byte, contentType string) {
	r.Body = body
	if len(body) != 0 {
		r.Headers.Set("Content-Length", fmt.Sprintf("%d", len(body)+1))
	} else {
		r.Headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}
	r.Headers.Set("Content-Type", contentType)
}

// Checkbody checks the length of the provided body
// according to the value of the [Content-Length] header,
// and returns error if any.
func (r *Response) CheckBody() error {
	if r.ContentLength == 0 {
		r.Body = []byte{}
	}
	if len(r.Body) != r.ContentLength {
		return errors.New("incomplete body received")
	}
	return nil
}

// AppendBody adds the remaining bytes of the message to the end of the body.
func (r *Response) AppendBody(data []byte) {
	r.Body = append(r.Body, data...)
}

// parseResponseLine parses the response line according to [RFC 9112 Section 4]
// provided as a string input, returns [*Response], number of characters consumed
// and error if any.
//
// If the string input is an incomplete header, (nil, 0, nil) is returned
// indicating that the function requires more input.
//
// The [RFC 9112 Section 4] describes response-line as follows:
//
//	status-line = HTTP-version SP status-code SP [ reason-phrase ]
//	HTTP-version = HTTP-name "/" DIGIT "." DIGIT
//	HTTP-name = %x48.54.54.50 ; HTTP
//	status-code = 3DIGIT
//	reason-phrase = 1*( HTAB / SP / VCHAR / obs-text )
//
// [RFC 9112 Section 4]: https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
func ParseResponseLine(s string) (*Response, int, error) {
	if len(s) == 0 {
		return nil, 0, errors.New("empty response line received")
	}
	idx := strings.Index(s, CRLFDELIMETER)
	if idx == -1 {
		return nil, 0, nil
	}
	line := s[:idx]
	consumedBytes := idx + len(CRLFDELIMETER)

	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return nil, consumedBytes, errors.New("invalid number of parts in response line")
	}
	version := parts[0]
	_, ver, found := strings.Cut(version, "/")
	if !found {
		return nil, consumedBytes, errors.New("no / in HTTP Version")
	}
	if ver != HTTP_VERSION {
		return nil, consumedBytes, errors.New("invalid HTTP version received")
	}

	sc, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, consumedBytes, errors.New("invalid Status Code received")
	}

	return &Response{
		ResponseLine: ResponseLine{
			HTTPVersion:  ver,
			StatusCode:   HTTPStatusCode(sc),
			ReasonPhrase: parts[2],
		},
		Headers: NewHeaders(),
		Body:    make([]byte, 0),
	}, consumedBytes, nil

}

type ResponseWriter struct {
	Writer io.Writer
}

func NewResponseWriter(w io.Writer) *ResponseWriter {
	return &ResponseWriter{
		Writer: w,
	}
}

func (w *ResponseWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

// WriteStatusLine builds and writes the status line based on the
// statusCode provided, returns error if any.
func (w *ResponseWriter) WriteStatusLine(statusCode HTTPStatusCode) error {

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

// WriteHeaders writes [HTTPHeaders] in http format,
// returns error if any.
//
// The http format is as follows:
//
//	field-name: value\r\n
func (w *ResponseWriter) WriteHeaders(headers HTTPHeaders) error {
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
func (w *ResponseWriter) WriteChunkedBody(p []byte) (int, error) {
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
func (w *ResponseWriter) WriteChunkedBodyDone() (int, error) {
	return w.Write([]byte("0\r\n\r\n"))

}

// GetDefaultHeaders creates new default headers and
// returns [HTTPHeaders] them based on the
// content length provided
//
// It creates the following headers:
//   - Content-Length: [contentLen]
//   - Connection: close
//   - Content-Type: text/html
func GetDefaultHeaders(contentLen int) HTTPHeaders {
	hdr := NewHeaders()

	contentlength := strconv.Itoa(contentLen)
	hdr.Set("Content-Length", contentlength)
	hdr.Set("Connection", "close")
	hdr.Set("Content-Type", "text/html")
	return hdr
}
