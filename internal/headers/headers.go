package headers

import (
	"errors"
	"strings"
)

const crlf string = "\r\n"

type HTTPHeaders struct {
	HeadersMap map[string]string
}

func NewHeaders() HTTPHeaders {
	return HTTPHeaders{
		HeadersMap: make(map[string]string),
	}
}

// Parse parses the header lines according to [RFC 9112 Section 5],
// provided as a byte array, returns number of bytes parsed, bool indicating
// parsing completion and error if any.
//
// [RFC 9112 Section 5]: https://datatracker.ietf.org/doc/html/rfc9112
func (h HTTPHeaders) Parse(data []byte) (int, bool, error) {
	if len(data) == 0 {
		return 0, false, errors.New("empty request line")
	}
	totalN := 0
	done := false

	for !done {
		var n int
		var err error
		n, done, err = h.parseHeaderLine(string(data))
		if err != nil {
			return 0, done, err
		}
		totalN += n
		data = data[n:]

		if n == 0 || len(data) == 0 {
			break
		}
	}
	return totalN, done, nil
}

// parseHeaderLine parses the header line according to [RFC 9112 Section 5]
// provided as a string input, returns number of characters parsed, bool indicating
// parsing completion and error if any.
//
// If the string input is an incomplete header, (0, false, nil) is returned
// indicating that the function requires more data.
//
// The [RFC 9112 5 field-line] is as follows:
//
//	field-line   = field-name ":" OWS field-value OWS
//
//	field-name     = token
//	token          = 1*tchar
//	tchar          = "!" / "#" / "$" / "%" / "&" / "'" / "*"
//		              / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
//		              / DIGIT / ALPHA
//		              ; any VCHAR, except delimiters
//
// The [RFC 9110 5.5 field-value] is as follows:
//
//	field-value    = *field-content
//	field-content  = field-vchar
//	                 [ 1*( SP / HTAB / field-vchar ) field-vchar ]
//	field-vchar    = VCHAR / obs-text
//	obs-text       = %x80-FF
//
// [RFC 9112 Section 5]: https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
// [RFC 9112 5 field-line]: https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
// [RFC 9110 5.5 field-value]: https://www.rfc-editor.org/rfc/rfc9110
func (h HTTPHeaders) parseHeaderLine(s string) (n int, done bool, err error) {
	if h.HeadersMap == nil {
		h.HeadersMap = make(map[string]string)
	}

	if len(s) == 0 {
		return 0, false, errors.New("empty header line received")
	}
	idx := strings.Index(s, crlf)
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 0, true, nil
	}

	consumed := idx + len(crlf)

	key, val, found := strings.Cut(string(s[:idx]), ":")
	if !found || len(key) == 0 || !isToken(key) {
		return 0, false, errors.New("malformed header received")
	}

	key = strings.ToLower(key)
	v, exists := h.HeadersMap[key]
	if exists {
		h.HeadersMap[key] = v + "," + strings.TrimSpace(val)
	} else {
		h.HeadersMap[key] = strings.TrimSpace(val)

	}

	return consumed, false, nil

}

// Get returns the value of a header by name.
func (h HTTPHeaders) Get(name string) string {
	val, exists := h.HeadersMap[strings.ToLower(name)]
	if !exists {
		return ""
	}
	return val
}

// Set sets or overwrites a header with the given value.
func (h HTTPHeaders) Set(name string, val string) {
	h.HeadersMap[strings.ToLower(name)] = val
}

// Replace updates the value of a header, assuming it already exists.
func (h HTTPHeaders) Replace(name string, val string) {
	h.HeadersMap[strings.ToLower(name)] = val
}

// Delete removes the header by the given header name.
func (h HTTPHeaders) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.HeadersMap, name)
}

// isToken checks whether a given string is Token.
//
// Tokens are short textual identifiers that do not include whitespace or delimiters.
//
// see [rfc 9110 5.6.2. Tokens]
//
// [rfc 9110 5.6.2. Tokens]: https://www.rfc-editor.org/rfc/rfc9110.html#name-tokens
func isToken(s string) bool {
	for _, c := range s {
		if !isTchar(c) {
			return false
		}
	}
	return true
}

// isAlphaNumeric checks whether a given rune is alphabet or a digit rune.
func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// isTchar checks whether a given rune is tchar.
//
// tchar are defined as a list of character except delimeter characters or alphanumeric characters.
//
// see [rfc 9110 5.6.2. Tokens]
//
// [rfc 9110 5.6.2. Tokens]: https://www.rfc-editor.org/rfc/rfc9110.html#name-tokens
func isTchar(c rune) bool {
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return isAlphaNumeric(c)
	}
}
