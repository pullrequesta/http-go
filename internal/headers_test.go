package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaderLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *HTTPHeaders
	}{
		{
			name:  "valid single header",
			input: "\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"": "",
				},
			},
		},
		{
			name:  "valid single header",
			input: "Host: localhost:42069\r\n\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "valid single header with extra whitespace",
			input: "Host:        localhost:42069\r\n\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "valid two headers",
			input: "Host: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"Host": "localhost:42069,localhost:42069",
				},
			},
		},
	}

	for _, tc := range testCases {
		hdr := NewHeaders()
		n, done, err := hdr.Parse([]byte((tc.input)))
		assert.NoError(t, err)
		fmt.Printf("number of bytes consumed by headers: %d\n", n)
		fmt.Println(done)
		assert.True(t, done)
		for key := range tc.expected.HeadersMap {
			assert.Equal(t, tc.expected.HeadersMap[key], hdr.Get(key))
		}
	}

}
func TestToTitleCase(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"host", "Host"},
		{"hOSt", "Host"},
		{"HosT", "Host"},
		{"Host", "Host"},
		{"x-auth-token", "X-Auth-Token"},
		{"x-auth-TOKEN", "X-Auth-Token"},
		{"content-length", "Content-Length"},
	}

	for _, tc := range testCases {
		got := titleCase(tc.input)
		assert.Equal(t, got, tc.expected)
	}

}

func TestParseHeaderLineReturnsError(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *HTTPHeaders
	}{
		{
			name:  "invalid single header",
			input: "       Host : localhost:42069       \r\n\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "invalid single header",
			input: "H©st: localhost:42069\r\n\r\n",
			expected: &HTTPHeaders{
				HeadersMap: map[string]string{
					"H©st": "localhost:42069",
				},
			},
		},
	}

	for _, tc := range testCases {
		hdr := NewHeaders()
		n, done, err := hdr.Parse([]byte((tc.input)))
		assert.Error(t, err)
		fmt.Printf("number of bytes consumed by headers: %d\n", n)
		fmt.Println(done)
		assert.False(t, done)
		for key := range tc.expected.HeadersMap {
			assert.NotEqual(t, tc.expected.HeadersMap[key], hdr.Get(key))
		}

	}

}
