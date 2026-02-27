package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

// isValidFieldNameChar returns true if the character is allowed in a field-name
func isValidFieldNameChar(c byte) bool {
	// Uppercase: A-Z
	if c >= 'A' && c <= 'Z' {
		return true
	}
	// Lowercase: a-z
	if c >= 'a' && c <= 'z' {
		return true
	}
	// Digits: 0-9
	if c >= '0' && c <= '9' {
		return true
	}
	// Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// isValidFieldName checks if a string contains only valid field-name characters
func isValidFieldName(name string) bool {
	for _, c := range name {
		if !isValidFieldNameChar(byte(c)) {
			return false
		}
	}
	return len(name) > 0
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := string(parts[0])

	// Check for trailing spaces (these are invalid)
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	// Trim leading spaces
	trimmedKey := strings.TrimLeft(key, " ")
	key = strings.ToLower(trimmedKey)

	// Validate that key contains only allowed characters
	if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	stringValue := string(value)

	exValue := h[key]
	if exValue != "" {
		stringValue = exValue + ", " + stringValue
	}
	h.Set(key, stringValue)

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	h[key] = value
}
func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func NewHeaders() Headers {
	return Headers{}
}
