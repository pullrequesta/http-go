package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestString(t *testing.T) {
	testCases := []struct {
		input    *Request
		expected []string
	}{
		{
			input: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/coffee",
					HttpVersion:   "1.1",
				},
				Headers: HTTPHeaders{
					HeadersMap: map[string]string{
						"Host":           "localhost",
						"Content-Length": "13",
					},
				},
				Body: []byte("hello world!"),
			},
			expected: []string{
				"GET /coffee HTTP/1.1\r\n",
				"Host: localhost\r\n",
				"Content-Length: 13\r\n",
				"\r\n\r\n",
				"hello world!\n"},
		},
	}
	for _, tc := range testCases {
		got := tc.input.String()
		for _, exp := range tc.expected {
			assert.Contains(t, got, exp)
		}
	}

}

func TestParseRequestLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *Request
	}{
		{
			name:  "good GET Request line",
			input: "GET /coffee HTTP/1.1\r\n",
			expected: &Request{
				RequestLine: RequestLine{
					Method:        "GET",
					RequestTarget: "/coffee",
					HttpVersion:   "1.1"},
				Headers: NewHeaders(),
				Body:    make([]byte, 0),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, n, err := ParseRequestLine(tc.input)
			fmt.Printf("number of bytes consumed by request line: %d\n", n)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.expected.RequestLine.Method, got.RequestLine.Method)
			assert.Equal(t, tc.expected.RequestLine.RequestTarget, got.RequestLine.RequestTarget)
			assert.Equal(t, tc.expected.RequestLine.HttpVersion, got.RequestLine.HttpVersion)
			assert.Equal(t, got, tc.expected)
		})

	}

}

func TestParseRequestLineReturnsError(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty request line wihtout crlf",
			input:    "",
			expected: "empty request line received",
		},
		{
			name:     "empty request line wiht crlf",
			input:    "\r\n",
			expected: "invalid number of parts in request line",
		},
		{
			name:     "incomplete request line with crlf",
			input:    "/coffee HTTP/\r\n",
			expected: "invalid number of parts in request line",
		},
		{
			name:     "invalid method in request line",
			input:    "get /coffee HTTP/\r\n",
			expected: "invalid request method received",
		},
		{
			name:     "invalid number of parts in request line",
			input:    "POST /coffee\r\n",
			expected: "invalid number of parts in request line",
		},
		{
			name:     "no separater '/' in HTTP version",
			input:    "POST /coffee HTTP1.1\r\n",
			expected: "no / in HTTP version",
		},
		{
			name:     "invalid HTTP Version in request line",
			input:    "POST /coffee HTTP/11\r\n",
			expected: "invalid HTTP version received",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, n, err := ParseRequestLine(tc.input)
			fmt.Printf("number of bytes consumed by request line: %d\n", n)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tc.expected)
			assert.Nil(t, got)
		})

	}
}
