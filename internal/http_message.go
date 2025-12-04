package internal

import (
	"errors"
	"io"
)

//RFC 9112 Section 2: (https://datatracker.ietf.org/doc/html/rfc9112#name-message-format)
// HTTP-message   = start-line CRLF
//
//	*( field-line CRLF )
//	CRLF
//	[ message-body ]
//
// start-line     = request-line / status-line
// field-line   = field-name ":" OWS field-value OWS
// message-body = *OCTET

const bufferSize int = 1024

type HTTPMessage interface {
	ParseHeaders([]byte) (int, bool, error)
	GetHeader(string) string
	GetContentLength() int
	SetContentLength(int)
	GetBody() []byte
	SetBody([]byte, string)
	CheckBody() error
	AppendBody([]byte)
}

// MessageFromReader parses the HTTPMessage from [io.Reader] and
// returns [HTTPMessage], error if any.
func MessageFromReader(reader io.Reader) (HTTPMessage, error) {
	buff := make([]byte, bufferSize)
	readToIndex := 0

	p := Parser{
		state:   ParserStateInitial,
		msgType: unknown,
	}

	for p.state != ParserStateDone {
		numBytesRead, err := reader.Read(buff[readToIndex:])
		if err == io.EOF {
			p.state = ParserStateDone
			break
		}
		readToIndex += numBytesRead
		if readToIndex == len(buff) {
			newbuffer := make([]byte, len(buff)*2)
			copy(newbuffer, buff)
			buff = newbuffer
		}
		numBytesParsed, err := p.Parse(buff[:readToIndex])
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

	switch p.msgType {
	case unknown:
		return nil, errors.New("unable to parse")
	case httpRequest:
		if err := p.req.CheckBody(); err != nil {
			return nil, err
		}
		return p.req, nil
	case httpResponse:
		if err := p.resp.CheckBody(); err != nil {
			return nil, err
		}
		return p.resp, nil

	}
	return nil, errors.New("invalid state reached")

}
