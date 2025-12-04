package internal

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
func TestRequest(t *testing.T) {
	testCases := []struct {
		name     string
		input    *chunkReader
		expected *Request
	}{
		{
			name: "Standard Body (valid)",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:42069\r\n" +
					"Content-Length: 13\r\n" +
					"\r\n" +
					"hello world!\n",
				numBytesPerRead: 7,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "POST",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "host:42069",
						"Content-Length": "13",
					},
				},
				ContentLength: 13,
				Body:          []byte("hello world!\n"),
			},
		},
		{
			name: "Empty Body, 0 reported content length (valid)",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:42069\r\n" +
					"Content-Length: 0\r\n" +
					"\r\n" +
					"",
				numBytesPerRead: 7,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "POST",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "host:42069",
						"Content-Length": "0",
					},
				},
				ContentLength: 0,
				Body:          []byte{},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := MessageFromReader(tc.input)
			r, ok := msg.(*Request)
			assert.True(t, ok)
			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, tc.expected.Body, r.Body)
			assert.Equal(t, tc.expected.ContentLength, r.ContentLength)
		})
	}
}

func TestRequestWithoutContentHeader(t *testing.T) {
	type ExpectedStr struct {
		r *Request
		h []string
	}
	testCases := []struct {
		name     string
		input    *chunkReader
		expected ExpectedStr
	}{
		{
			name: "Empty Body, no reported content length (valid)",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:42069\r\n" +
					"\r\n" +
					"",
				numBytesPerRead: 7,
			},
			expected: ExpectedStr{
				r: &Request{
					RequestLine: RequestLine{
						Method:        "POST",
						RequestTarget: "/",
						HttpVersion:   "1.1",
					},
					Headers: HTTPHeaders{
						HeadersMap: map[string]string{
							"Host":           "host:42069",
							"Content-Length": "0",
						},
					},
					ContentLength: 0,
					Body:          []byte{},
				},
				h: []string{"Host: host:42069", "Content-Length: 0"},
			},
		},
		{
			name: "No Content-Length but Body Exists",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:42069\r\n" +
					"\r\n" +
					"partial",
				numBytesPerRead: 7,
			},
			expected: ExpectedStr{
				r: &Request{
					RequestLine: RequestLine{
						Method:        "POST",
						RequestTarget: "/",
						HttpVersion:   "1.1",
					},
					Headers: HTTPHeaders{
						HeadersMap: map[string]string{
							"Host":           "host:42069",
							"Content-Length": "0",
						},
					},
					Body: []byte{},
				},
				h: []string{"Host: host:42069", "Content-Length: 0"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := MessageFromReader(tc.input)
			r, ok := msg.(*Request)
			assert.True(t, ok)
			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, tc.expected.r.Body, r.Body)
			assert.Equal(t, tc.expected.r.ContentLength, r.ContentLength)
			for _, h := range tc.expected.h {
				assert.Contains(t, r.String(), h)
			}

		})
	}
}

func TestRequestReturnsError(t *testing.T) {
	testCases := []struct {
		name  string
		input *chunkReader
	}{
		{
			name: "Bad GET Request line with a malformed header",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nHost localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
		},
		{
			name: "Bad GET Request line with a malformed header",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nHost : localhost:42069       \r\n\r\n",
				numBytesPerRead: 3,
			},
		},
		{
			name: "Bad GET Request line with invalid character",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nH©st: localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
		},
		{
			name: "Empty Request line",
			input: &chunkReader{
				data:            "",
				numBytesPerRead: 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := MessageFromReader(tc.input)
			assert.Error(t, err)
			assert.Nil(t, msg)
		})
	}

}

func TestResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    *chunkReader
		expected *Response
	}{
		{
			name: "Good GET ResponseLine line with a single header",
			input: &chunkReader{
				data:            "HTTP/1.1 400 BAD Request\r\nHost: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Response{
				ResponseLine: ResponseLine{
					HTTPVersion:  "1.1",
					StatusCode:   400,
					ReasonPhrase: "BAD Request",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069,localhost:42069",
					},
				},
			},
		},
		{
			name: "Good GET ResponseLine line with a single header",
			input: &chunkReader{
				data:            "HTTP/1.1 200 OK\r\nHost: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Response{
				ResponseLine: ResponseLine{
					HTTPVersion:  "1.1",
					StatusCode:   200,
					ReasonPhrase: "OK",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069,localhost:42069",
					},
				},
			},
		},
		{
			name: "Standard Body (valid)",
			input: &chunkReader{
				data: "HTTP/1.1 200 OK\r\n" +
					"Host: host:42069\r\n" +
					"Content-Length: 13\r\n" +
					"\r\n" +
					"hello world!\n",
				numBytesPerRead: 7,
			},
			expected: &Response{
				ResponseLine: ResponseLine{
					HTTPVersion:  "1.1",
					StatusCode:   200,
					ReasonPhrase: "OK",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "host:42069",
						"Content-Length": "13",
					},
				},
				ContentLength: 13,
				Body:          []byte("hello world!\n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := MessageFromReader(tc.input)
			r, ok := msg.(*Response)
			assert.True(t, ok)
			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, tc.expected.ResponseLine.HTTPVersion, r.ResponseLine.HTTPVersion)
			assert.Equal(t, tc.expected.ResponseLine.ReasonPhrase, r.ResponseLine.ReasonPhrase)
			assert.Equal(t, tc.expected.ResponseLine.StatusCode, r.ResponseLine.StatusCode)
			for key := range tc.expected.Headers.HeadersMap {
				assert.Equal(t, tc.expected.Headers.HeadersMap[key], r.Headers.Get(key))
			}
		})
	}
}
