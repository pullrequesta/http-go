package internal

import (
	"errors"
	"strconv"
	"strings"
)

type ParserState int

const (
	ParserStateInitial ParserState = iota + 1 // parser state just started or waiting for input
	ParserStateHeader                         // parser state for parsing the headers
	ParserStateBody                           // parser state for parsing the body
	ParserStateDone                           // parser state for finishing the http message
)

type httpMessagetype int

const (
	unknown httpMessagetype = iota
	httpRequest
	httpResponse
)

type Parser struct {
	state   ParserState
	req     *Request
	resp    *Response
	msgType httpMessagetype
}

// Parse accepts the next slice of bytes that needs to be parsed.
// It updates the state of the parser.
//
// It takes data as a []byte input, returns number of bytes it parsed
// and error if any.
func (p *Parser) Parse(data []byte) (int, error) {
	totalBytesParsed := 0
	if len(data) == 0 {
		return 0, errors.New("empty request line received")
	}

	for p.state != ParserStateDone {
		n, err := p.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			// end of headers
			// consume \r\n
			if p.state != ParserStateBody {
				break
			}
			n = len(CRLFDELIMETER)
			switch p.msgType {
			case unknown:
				break
			case httpRequest:
				contentLengthStr := p.req.GetHeader("Content-Length")
				if contentLengthStr == "" {
					p.req.SetBody([]byte{}, "")
					p.state = ParserStateDone
					return len(data), nil
				}
				contentLen, err := strconv.Atoi(contentLengthStr)
				if err != nil {
					return 0, errors.New("invalid content-length received")
				}
				p.req.SetContentLength(contentLen)
			case httpResponse:
				contentLengthStr := p.resp.GetHeader("Content-Length")
				if contentLengthStr == "" {
					p.resp.SetBody([]byte{}, "")
					p.state = ParserStateDone
					return len(data), nil
				}
				contentLen, err := strconv.Atoi(contentLengthStr)
				if err != nil {
					return 0, errors.New("invalid content-length received")
				}
				p.resp.SetContentLength(contentLen)
			}

		}
		totalBytesParsed += n
		if totalBytesParsed >= len(data) {
			break
		}
	}
	return totalBytesParsed, nil
}

func (p *Parser) IsInInvalidState() bool {
	return p.state != ParserStateInitial && p.msgType == unknown
}

// parseSingle implements the switch/case logic that
// handles the actual parsing of the HTTPMessage based on the the [ParserState].
//
// It takes data as a []byte input, returns number of bytes parsed
// and error if any.
//
// [ParserStateInitial] parses the request line to parseRequestLine function
// and sets the state to parseHeaderState.
//
// [ParserStateHeader] parses multiple headers in r.Headers.Parse function
// and sets the state to parseBodyState.
//
// [ParserStateBody] appends the body according to the Content-Length
// and sets the state to [ParserStateDone].
func (p *Parser) parseSingle(data []byte) (int, error) {

	if p.IsInInvalidState() {
		return 0, errors.New("neither request nor response and not initial state")
	}

state_switch:
	switch p.state {

	case ParserStateInitial:
		switch p.msgType {
		case unknown:
			var err error
			p.req, p.resp, err = parseFirstLine(string(data))
			if err != nil {
				return 0, err
			}
			if p.req == nil && p.resp == nil {
				return 0, nil
			}
			if p.req != nil {
				p.msgType = httpRequest
			} else {
				p.msgType = httpResponse
			}
			goto state_switch
		case httpRequest:
			r, n, err := ParseRequestLine(string(data))
			if err != nil {
				return 0, err
			}
			p.req = r
			if n == 0 {
				return 0, nil
			}
			p.state = ParserStateHeader
			return n, nil
		case httpResponse:
			r, n, err := ParseResponseLine(string(data))
			if err != nil {
				return 0, err
			}
			p.resp = r
			if n == 0 {
				return 0, nil
			}
			p.state = ParserStateHeader
			return n, nil
		}

	case ParserStateHeader:
		var n int
		var ok bool
		var err error
		switch p.msgType {
		case httpRequest:
			n, ok, err = p.req.ParseHeaders(data)
		case httpResponse:
			n, ok, err = p.resp.ParseHeaders(data)
		}

		if err != nil {
			return 0, err
		}
		if n == 0 {
			if !ok {
				return n, nil
			}
			p.state = ParserStateBody
		}
		return n, nil

	case ParserStateBody:
		var bodyLen int
		var cl int
		switch p.msgType {
		case httpRequest:
			p.req.AppendBody(data)
			bodyLen = len(p.req.GetBody())
			cl = p.req.GetContentLength()
		case httpResponse:
			p.resp.AppendBody(data)
			bodyLen = len(p.resp.GetBody())
			cl = p.resp.GetContentLength()
		}

		if bodyLen > cl {
			return 0, errors.New("body exceeds the declared content-length")
		}

		if bodyLen == cl {
			p.state = ParserStateDone
		}

		return len(data), nil
	case ParserStateDone:
		return 0, errors.New("trying to read data in a done state")

	default:
		return 0, errors.New("unknown state")

	}
	return 0, errors.New("unknown state2")

}

func parseFirstLine(s string) (*Request, *Response, error) {
	if len(s) == 0 {
		return nil, nil, errors.New("empty request line received")
	}

	if !strings.Contains(s, CRLFDELIMETER) {
		return nil, nil, nil
	}

	if strings.HasPrefix(s, "HTTP/1.1") {
		return nil, &Response{}, nil
	}

	return &Request{}, nil, nil
}
