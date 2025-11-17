package request

import (
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const httpVersion string = "1.1"
const crlf string = "\r\n"
const bufferSize int = 1024

// Define an enum-like type for parser state
type ParserState int

const (
	initialState     ParserState = iota + 1 // parser state just started or waiting for input
	parseHeaderState                        // parser state for parsing the headers
	parseBodyState                          // parser state for parsing the body
	doneState                               // parser state for finishing the http message
)

// Define the request struct for Http message
type Request struct {
	RequestLine   RequestLine
	Headers       headers.HTTPHeaders
	ContentLength int
	Body          []byte
	State         ParserState
}

// Define the request-line struct for Http message
type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type errReqLine struct {
	Err   error
	Value string
}

func (e *errReqLine) Error() string {
	return "Err: " + e.Err.Error() + ", for value [" + e.Value + "]"
}

func NewRequest() *Request {
	return &Request{
		State:   initialState,
		Headers: headers.NewHeaders(),
	}
}

func (r *Request) GetHTTPVersion() string {
	return r.RequestLine.HttpVersion
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, bufferSize)
	readToIndex := 0

	r := NewRequest()

	for r.State != doneState {
		// read into the buffer
		numBytesRead, err := reader.Read(buff[readToIndex:])
		if err == io.EOF {
			r.State = doneState
			break
		}
		readToIndex += numBytesRead
		if readToIndex == len(buff) {
			newbuffer := make([]byte, len(buff)*2)
			copy(newbuffer, buff)
			buff = newbuffer
		}
		// parse from the buffer
		numBytesParsed, err := r.parse(buff[:readToIndex])
		if err != nil {
			return nil, err
		}
		if numBytesParsed > 0 {
			newbuff := make([]byte, len(buff))
			copy(newbuff, buff[numBytesParsed:])
			buff = newbuff
			readToIndex -= numBytesParsed
		}
	}
	if r.ContentLength == 0 {
		r.Body = []byte{}
	}
	if len(r.Body) != r.ContentLength {
		return nil, errors.New("incomplete body received")
	}
	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {

	totalBytesParsed := 0
	if len(data) == 0 {
		return 0, errors.New("empty request line received")
	}

	for r.State != doneState {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err

		}
		if n == 0 {
			// end of headers
			// consume \r\n
			if r.State != parseBodyState {
				break
			}
			n = len(crlf)
			contentLengthStr := r.Headers.Get("Content-Length")
			if contentLengthStr == "" {
				r.Body = []byte{}
				r.State = doneState
				return len(data), nil
			}
			contentLen, err := strconv.Atoi(contentLengthStr)
			if err != nil {
				return 0, errors.New("invalid content-length received")
			}
			r.ContentLength = contentLen

		}
		totalBytesParsed += n
		if totalBytesParsed >= len(data) {
			break
		}
	}

	return totalBytesParsed, nil

}
func (r *Request) parseSingle(data []byte) (int, error) {

	switch r.State {

	case initialState:
		reqLine, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.State = parseHeaderState
		// fmt.Printf("number of bytes consumed by the request line: %d\n", n)
		return n, nil
	case parseHeaderState:
		n, ok, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			if !ok {
				return n, nil
			}
			r.State = parseBodyState
		}
		// fmt.Printf("number of bytes consumed by the headers: %d\n", n)
		return n, nil
	case parseBodyState:
		r.Body = append(r.Body, data...)
		bodyLen := len(r.Body)

		if bodyLen > r.ContentLength {
			return 0, errors.New("body exceeds the declared content-length")
		}

		if bodyLen == r.ContentLength {
			r.State = doneState
		}

		// fmt.Printf("number of bytes consumed by the body: %d\n", len(data))
		return len(data), nil

	case doneState:
		return 0, errors.New("trying to read data in a done state")

	default:
		return 0, errors.New("unknown state")

	}
}

func parseRequestLine(s string) (*RequestLine, int, error) {
	if len(s) == 0 {
		return nil, 0, errors.New("empty request line received")
	}

	idx := strings.Index(s, crlf)
	if idx == -1 {
		return nil, 0, nil
	}
	line := s[:idx]
	consumedBytes := idx + len(crlf)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, consumedBytes, errors.New("invalid request line received")
	}
	method := parts[0]
	if !isUpper(method) {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid request method received"), Value: method}
	}
	version := parts[2]
	_, ver, found := strings.Cut(version, "/")
	if !found {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid Http version format received"), Value: version}
	}
	if ver != httpVersion {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid Http version received"), Value: ver}
	}

	return &RequestLine{Method: method,
		RequestTarget: parts[1],
		HttpVersion:   ver,
	}, consumedBytes, nil

}

func isUpper(m string) bool {

	for _, v := range m {
		if !unicode.IsUpper(v) && unicode.IsLetter(v) {
			return false
		}
	}
	return true
}
