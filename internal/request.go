package internal

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type errReqLine struct {
	Err   error
	Value string
}

func (e *errReqLine) Error() string {
	return "Err: " + e.Err.Error() + ", for value [" + e.Value + "]"
}

type Request struct {
	RequestLine   RequestLine
	Headers       HTTPHeaders
	ContentLength int
	Body          []byte
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

func newRequestLine(method, target string) *RequestLine {
	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   HTTP_VERSION,
	}
}

func NewRequest(method string, target string) *Request {
	return &Request{
		RequestLine: *newRequestLine(method, target),
		Headers:     NewHeaders(),
	}
}

func NewUnparsedRequest() *Request {
	return &Request{
		Headers: NewHeaders(),
	}
}

func (r *RequestLine) String() string {
	var b strings.Builder
	if r.Method != "" {
		b.WriteString(r.Method + " ")
	}
	if r.RequestTarget != "" {
		b.WriteString(r.RequestTarget + " ")
	}
	if r.HttpVersion != "" {
		b.WriteString("HTTP/" + r.HttpVersion + CRLFDELIMETER)
	}
	return b.String()
}

func (r *Request) String() string {

	builder := strings.Builder{}

	builder.WriteString(r.RequestLine.String())

	if r.Headers.HeadersMap != nil {
		for k, v := range r.Headers.HeadersMap {
			builder.WriteString(k + ": " + v + CRLFDELIMETER)
		}
	}

	builder.WriteString(CRLFDELIMETER)

	if len(r.Body) != 0 {
		builder.WriteString(string(r.Body) + NEWLINE)
	}
	return builder.String()

}

// ParseHeaders parses data in the ParseHeaders function as an input,
// returns number of bytes parsed, bool indicating
// parsing completion and error if any.
func (r *Request) ParseHeaders(data []byte) (int, bool, error) {
	return r.Headers.Parse(data)
}

// GetHeader returns the value of a particular header by its name.
func (r *Request) GetHeader(name string) string {
	return r.Headers.Get(name)
}

// GetContentLength returns the [Content-Length] in the form of an integer.
func (r *Request) GetContentLength() int {
	return r.ContentLength
}

// SetContentLength sets the [Content-Length] field of the header.
func (r *Request) SetContentLength(cl int) {
	r.ContentLength = cl
}

// GetBody returns the body of the request.
func (r *Request) GetBody() []byte {
	return r.Body
}

// SetBody sets the body of the request,
// sets the [Content-Length] according the length of body,
// and sets the [Content-Type] to its values.
func (r *Request) SetBody(body []byte, contentType string) {

	r.Body = body
	if len(body) != 0 {
		r.Headers.Set("Content-Length", fmt.Sprintf("%d", len(body)+1))
	} else {
		r.Headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}

	if contentType != "" {
		r.Headers.Set("Content-Type", contentType)
	}
}

// Checkbody checks the length of the provided body
// according to the value of the [Content-Length] header,
// and returns error if any.
func (r *Request) CheckBody() error {
	if len(r.Body) != r.ContentLength {
		return errors.New("incomplete body received")
	}
	return nil
}

// AppendBody adds the remaining bytes of the message to the end of the body.
func (r *Request) AppendBody(data []byte) {
	r.Body = append(r.Body, data...)
}

// parseRequestLine parses the request line according to [RFC 9112 Section 3.1]
// provided as a string input, returns [*Request], number of characters consumed
// and error if any.
//
// If the string input is an incomplete header, (nil, 0, nil) is returned
// indicating that the function requires more input.
//
// The [RFC 9112 Section 3.1] describes request-line as follows:
//
//	request-line   = method SP request-target SP HTTP-version
//	method         = token
//	request-target = origin-form
//	               / absolute-form
//	               / authority-form
//	               / asterisk-form
//	HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
//	HTTP-name     = %s"HTTP"
//
// [RFC 9112 Section 3.1]: https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
func ParseRequestLine(s string) (*Request, int, error) {
	if len(s) == 0 {
		return nil, 0, errors.New("empty request line received")
	}

	idx := strings.Index(s, CRLFDELIMETER)
	if idx == -1 {
		return nil, 0, nil
	}
	line := s[:idx]
	consumedBytes := idx + len(CRLFDELIMETER)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, consumedBytes, errors.New("invalid number of parts in request line")
	}
	method := parts[0]
	if !isMethod(method) {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid request method received"), Value: method}
	}
	version := parts[2]
	_, ver, found := strings.Cut(version, "/")
	if !found {
		return nil, consumedBytes, &errReqLine{Err: errors.New("no / in HTTP version"), Value: version}
	}
	if ver != HTTP_VERSION {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid HTTP version received"), Value: ver}
	}

	return &Request{
		RequestLine: RequestLine{
			Method:        method,
			RequestTarget: parts[1],
			HttpVersion:   ver,
		},
		Headers: NewHeaders(),
		Body:    make([]byte, 0),
	}, consumedBytes, nil

}

// isMethod checks whether the given request method is case-sensitive and alphabetic.
func isMethod(m string) bool {
	for _, v := range m {
		if !unicode.IsUpper(v) && unicode.IsLetter(v) {
			return false
		}
	}
	return true
}
