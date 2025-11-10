package headers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaderLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *Headers
	}{
		{
			name:  "valid single header",
			input: "\r\n",
			expected: &Headers{
				Headers: map[string]string{
					"": "",
				},
			},
		},
		{
			name:  "valid single header",
			input: "Host: localhost:42069\r\n\r\n",
			expected: &Headers{
				Headers: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "valid single header with extra whitespace",
			input: "Host:        localhost:42069\r\n\r\n",
			expected: &Headers{
				Headers: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "valid two headers",
			input: "Host: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
			expected: &Headers{
				Headers: map[string]string{
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
		for key := range tc.expected.Headers {
			assert.Equal(t, tc.expected.Headers[key], hdr.Get(key))
		}

	}

}

func TestParseHeaderLineReturnsError(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *Headers
	}{
		{
			name:  "invalid single header",
			input: "       Host : localhost:42069       \r\n\r\n",
			expected: &Headers{
				Headers: map[string]string{
					"Host": "localhost:42069",
				},
			},
		},
		{
			name:  "invalid single header",
			input: "H©st: localhost:42069\r\n\r\n",
			expected: &Headers{
				Headers: map[string]string{
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
		for key := range tc.expected.Headers {
			assert.NotEqual(t, tc.expected.Headers[key], hdr.Get(key))
		}

	}

}
