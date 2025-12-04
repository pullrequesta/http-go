package internal

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetBody(t *testing.T) {

	testCases := []struct {
		input       *Response
		contentType string
	}{
		{input: &Response{
			ResponseLine: ResponseLine{
				HTTPVersion:  "1.1",
				StatusCode:   200,
				ReasonPhrase: "OK",
			},
			Headers: HTTPHeaders{
				HeadersMap: map[string]string{
					"Content-Length": "13",
					"Content-Type":   "plain/text",
				},
			},
			Body: []byte("Hello World\n"),
		}, contentType: "plain/text",
		},
	}
	for _, tc := range testCases {
		tc.input.SetBody([]byte(tc.input.Body), tc.contentType)

	}

}

func TestParseResponseLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected *Response
	}{
		{
			input: "HTTP/1.1 200 OK\r\n",
			expected: &Response{
				ResponseLine: ResponseLine{
					HTTPVersion:  "1.1",
					StatusCode:   StatusOK,
					ReasonPhrase: "OK",
				},
				Headers: NewHeaders(),
				Body:    make([]byte, 0),
			},
		},
	}
	for _, tc := range testCases {
		r, _, err := ParseResponseLine(tc.input)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, tc.expected.ResponseLine.HTTPVersion, r.ResponseLine.HTTPVersion)
		assert.Equal(t, tc.expected.ResponseLine.StatusCode, r.ResponseLine.StatusCode)
		assert.Equal(t, tc.expected.ResponseLine.ReasonPhrase, r.ResponseLine.ReasonPhrase)
		assert.Equal(t, tc.expected.Headers, r.Headers)
		assert.Equal(t, tc.expected.Body, r.Body)

	}
}

func TestParseResponseLineReturnsError(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty response line wihtout crlf",
			input:    "",
			expected: "empty response line received",
		},
		{
			name:     "empty response line wiht crlf",
			input:    "\r\n",
			expected: "invalid number of parts in response line",
		},
		{
			name:     "incomplete response line with crlf",
			input:    "HTTP/1.1 200\r\n",
			expected: "invalid number of parts in response line",
		},
		{
			name:     "invalid number of parts in response line",
			input:    "HTTP/1.1 OK\r\n",
			expected: "invalid number of parts in response line",
		},
		{
			name:     "no separater '/' in HTTP version",
			input:    "HTTP1.1 200 OK\r\n",
			expected: "no / in HTTP version",
		},
		{
			name:     "invalid HTTP Version in response line",
			input:    "HTTP/11 200 OK\r\n",
			expected: "invalid HTTP version received",
		},
		{
			name:     "invalid status code in response line",
			input:    "HTTP/1.1 Have OK\r\n",
			expected: "invalid Status Code received",
		},
	}
	for _, tc := range testCases {
		r, _, err := ParseResponseLine(tc.input)
		assert.Nil(t, r)
		assert.Error(t, err)

	}
}

func TestWriteStatusLine(t *testing.T) {
	buff := bytes.Buffer{}
	respWriter := NewResponseWriter(&buff)

	testCases := []struct {
		input HTTPStatusCode
	}{
		{200},
		{400},
		{500},
		{21},
	}

	for _, tc := range testCases {
		err := respWriter.WriteStatusLine(tc.input)
		assert.NoError(t, err)
	}

}

func TestWriteHeaders(t *testing.T) {
	buff := bytes.Buffer{}
	respWriter := NewResponseWriter(&buff)
	testCases := []struct {
		input map[string]string
	}{
		{input: map[string]string{"Host": "localhost:42069"}},
		{input: map[string]string{"Content-Type": "text/plain"}},
	}

	for _, tc := range testCases {
		err := respWriter.WriteHeaders(HTTPHeaders{tc.input})
		assert.NoError(t, err)
	}

}

func TestWriteChunkedBody(t *testing.T) {
	buff := bytes.Buffer{}
	respWriter := NewResponseWriter(&buff)
	testCases := []struct {
		input string
	}{
		{"Welcome"},
		{"HelloWorld"},
	}
	for _, tc := range testCases {
		_, err := respWriter.WriteChunkedBody([]byte(tc.input))
		assert.NoError(t, err)
	}
	_, err := respWriter.WriteChunkedBodyDone()
	assert.NoError(t, err)

}
func TestGetDefaultHeaders(t *testing.T) {
	contentLen := []int{13, 23, 7}

	for _, cl := range contentLen {
		hdr := GetDefaultHeaders(cl)
		assert.NotNil(t, hdr)
		assert.Equal(t, fmt.Sprintf("%d", cl), hdr.Get("Content-Length"))
		assert.Equal(t, "text/html", hdr.Get("Content-Type"))
		assert.Equal(t, "close", hdr.Get("Connection"))
	}

}
