package request

import (
	"httpfromtcp/internal/headers"
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
func TestRequestLineWithHeadersAndBody(t *testing.T) {
	testCases := []struct {
		name     string
		input    *chunkReader
		expected *Request
	}{
		{
			name: "Standard Body (valid)",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:420\r\n" +
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
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "host:420",
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
					"Host: host:420\r\n" +
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
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "host:420",
						"Content-Length": "0",
					},
				},
				ContentLength: 0,
				Body:          []byte{},
			},
		},
		{
			name: "Empty Body, no reported content length (valid)",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:420\r\n" +
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
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "host:420",
					},
				},
				Body: []byte{},
			},
		},
		{
			name: "No Content-Length but Body Exists",
			input: &chunkReader{
				data: "POST / HTTP/1.1\r\n" +
					"Host: host:420\r\n" +
					"\r\n" +
					"partial",
				numBytesPerRead: 7,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "POST",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "host:420",
					},
				},
				Body: []byte{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := RequestFromReader(tc.input)
			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, tc.expected.Body, r.Body)
			assert.Equal(t, tc.expected.ContentLength, r.ContentLength)
		})
	}

}

// func TestRequestLineWithHeadersAndBodyReturnsError(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		input    *chunkReader
// 		expected *Request
// 	}{
// 		{
// 			name: "Body shorter than reported content length" (should error)",
// 			input: &chunkReader{
// 				data: "POST /submit HTTP/1.1\r\n" +
// 					"Host: localhost:42069\r\n" +
// 					"Content-Length: 20\r\n" +
// 					"\r\n" +
// 					"partial content",
// 				numBytesPerRead: 3,
// 			},
// 			expected: &Request{
// 				RequestLine: RequestLine{
// 					Method:        "POST",
// 					RequestTarget: "/submit",
// 					HttpVersion:   "1.1",
// 				},
// 				Headers: headers.Headers{
// 					Headers: map[string]string{
// 						"Host":           "localhost:42069",
// 						"Content-Length": "20",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			r, err := RequestFromReader(tc.input)
// 			assert.Nil(t, r)
// 			assert.Error(t, err)
// 		})
// 	}

// }

func TestRequestWithHeaders(t *testing.T) {
	testCases := []struct {
		name     string
		input    *chunkReader
		expected *Request
	}{
		{
			name: "good GET Request line with a single header",
			input: &chunkReader{
				data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069,localhost:42069",
					},
				},
			},
		},
		{
			name: "good GET Request line with a multiple headers",
			input: &chunkReader{
				data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":       "localhost:42069",
						"User-Agent": "curl/7.81.0",
						"Accept":     "*/*",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := RequestFromReader(tc.input)
			assert.NoError(t, err)
			assert.NotNil(t, r)

			for key := range tc.expected.Headers.HeadersMap {
				assert.Equal(t, tc.expected.Headers.HeadersMap[key], r.Headers.Get(key))
			}
			assert.Equal(t, tc.expected.RequestLine.Method, r.RequestLine.Method)
			assert.Equal(t, tc.expected.RequestLine.RequestTarget, r.RequestLine.RequestTarget)
			assert.Equal(t, tc.expected.RequestLine.HttpVersion, r.RequestLine.HttpVersion)
		})
	}
}

func TestRequestWithHeadersReturnsError(t *testing.T) {
	testCases := []struct {
		name     string
		input    *chunkReader
		expected *Request
	}{
		{
			name: "bad GET Request line with a malformed header",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nHost localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069",
					},
				},
			},
		},
		{
			name: "bad GET Request line with a malformed header",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nHost : localhost:42069       \r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069",
					},
				},
			},
		},
		{
			name: "bad GET Request line with a malformed header including invalid character",
			input: &chunkReader{
				data:            "GET / HTTP/11\r\nH©st: localhost:42069\r\n\r\n",
				numBytesPerRead: 3,
			},
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/",
					HttpVersion:   "1.1",
				},
				Headers: headers.HTTPHeaders{
					HeadersMap: map[string]string{
						"Host": "localhost:42069",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := RequestFromReader(tc.input)
			assert.Error(t, err)
			assert.Nil(t, r)
		})
	}

}

// func TestParseRequestLine(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		input    string
// 		expected *RequestLine
// 	}{
// 		{
// 			name:  "good GET Request line with pathinput",
// 			input: "GET /coffee HTTP/1.1",
// 			expected: &RequestLine{
// 				Method:        "GET",
// 				RequestTarget: "/coffee",
// 				HttpVersion:   "1.1",
// 			},
// 		},
// 		{
// 			name:  "good GET Request line",
// 			input: "POST / HTTP/1.1",
// 			expected: &RequestLine{
// 				Method:        "POST",
// 				RequestTarget: "/",
// 				HttpVersion:   "1.1",
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			got, n, err := parseRequestLine(tc.input)
// 			fmt.Printf("number of bytes consumed by request line: %d\n", n)
// 			assert.NoError(t, err)
// 			assert.NotNil(t, got)
// 			assert.Equal(t, tc.expected.Method, got.Method)
// 			assert.Equal(t, tc.expected.RequestTarget, got.RequestTarget)
// 			assert.Equal(t, tc.expected.HttpVersion, got.HttpVersion)
// 		})

// 	}

// }

// func TestParseRequestLineError(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		input    string
// 		expected *RequestLine
// 	}{
// 		{
// 			name:  "empty request line",
// 			input: "",
// 		},
// 		{
// 			name:  "invalid number of parts in request line",
// 			input: "/coffee HTTP/",
// 			expected: &RequestLine{
// 				Method:        "",
// 				RequestTarget: "/coffee",
// 				HttpVersion:   "",
// 			},
// 		},
// 		{
// 			name:  "invalid method in request line",
// 			input: "get /coffee HTTP/",
// 		},

// 		{
// 			name:  "invalid Http version in request line",
// 			input: "POST /coffee HTTP/2.0",
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			got, n, err := parseRequestLine(tc.input)
// 			fmt.Printf("number of bytes consumed by request line: %d\n", n)
// 			assert.Error(t, err)
// 			assert.Nil(t, got)
// 		})

// 	}
// }
