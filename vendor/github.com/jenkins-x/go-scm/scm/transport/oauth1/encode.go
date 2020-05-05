// Copyright (c) 2015 Dalton Hubble. All rights reserved.
// Copyrights licensed under the MIT License.

package oauth1

import (
	"bytes"
	"fmt"
	"strings"
)

// encodeParameterString encodes collected OAuth parameters
// into a parameter string as defined in RFC 5894 3.4.1.3.2.
func encodeParameterString(params map[string]string) string {
	return strings.Join(sortParameters(
		encodeParameters(params), "%s=%s"), "&")
}

// encodeParameters percent encodes parameter keys and
// values according to RFC5849 3.6 and RFC3986 2.1 and
// returns a new map.
func encodeParameters(params map[string]string) map[string]string {
	encoded := map[string]string{}
	for key, value := range params {
		encoded[percentEncode(key)] = percentEncode(value)
	}
	return encoded
}

// percentEncode percent encodes a string according to
// RFC 3986 2.1.
func percentEncode(input string) string {
	var buf bytes.Buffer
	for _, b := range []byte(input) {
		// if in unreserved set
		if shouldEscape(b) {
			buf.Write([]byte(fmt.Sprintf("%%%02X", b)))
		} else {
			// do not escape, write byte as-is
			buf.WriteByte(b)
		}
	}
	return buf.String()
}

// shouldEscape returns false if the byte is an unreserved
// character that should not be escaped and true otherwise,
// according to RFC 3986 2.1.
func shouldEscape(c byte) bool {
	// RFC3986 2.3 unreserved characters
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '.', '_', '~':
		return false
	}
	// all other bytes must be escaped
	return true
}
