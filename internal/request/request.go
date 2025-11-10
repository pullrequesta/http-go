package request

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log"
	"strings"
	"unicode"
)

const httpVersion string = "1.1"
const crlf string = "\r\n"
const bufferSize int = 1024

// Define an enum-like type for parser state
type State int

const (
	initializedState   State = iota + 1 // parser state just started or waiting for input
	parsingHeaderState                  // parser state for parsing the request
	doneState                           // parser state for parsing the request
)

// Define the request struct for Http message
type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       State
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

// Must helper function
func MUST[T any](arg T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return arg
}
func NewRequest() *Request {
	return &Request{
		State:   initializedState,
		Headers: headers.NewHeaders(),
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	//data := io.ReadAll(reader)
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

	return r, nil
}

// method for parsing request data into request line.
func (r *Request) parse(data []byte) (int, error) {

	totalBytesParsed := 0

	if len(data) == 0 {
		return 0, errors.New("empty request line")
	}

	for r.State != doneState {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			return totalBytesParsed, nil
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

	case initializedState:
		reqLine, n, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.State = parsingHeaderState
		fmt.Printf("number of bytes consumed by request line: %d\n", n)
		return n, nil
	case parsingHeaderState:
		n, ok, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 || !ok {
			return n, nil
		}
		r.State = doneState
		fmt.Printf("number of bytes consumed by headers: %d\n", n)
		return n, nil

	case doneState:
		return 0, errors.New("trying to read data in a done state")

	default:
		return 0, errors.New("unknown state")

	}
}

func parseRequestLine(s string) (*RequestLine, int, error) {
	if len(s) == 0 {
		return nil, 0, errors.New("empty request line")
	}

	idx := strings.Index(s, crlf)
	if idx == -1 {
		return nil, 0, nil
	}
	line := s[:idx]
	consumedBytes := idx + len(crlf)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, consumedBytes, errors.New("invalid request line")
	}
	method := parts[0]
	if !isUpper(method) {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid request method"), Value: method}
	}
	version := parts[2]
	_, ver, ok := strings.Cut(version, "/")
	if !ok {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid Http version format"), Value: ver}
	}
	if ver != httpVersion {
		return nil, consumedBytes, &errReqLine{Err: errors.New("invalid Http version format"), Value: ver}
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
