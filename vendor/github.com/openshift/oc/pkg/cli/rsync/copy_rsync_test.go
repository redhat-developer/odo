package rsync

import (
	"testing"
)

// TestRsyncEscapeCommand ensures that command line options supplied to oc rsync
// are properly escaped.
func TestRsyncEscapeCommand(t *testing.T) {
	// the strings to run through rsyncEscapeCommand
	stringsToEscape := []string{
		`thisshouldnotgetescapedorquoted`,
		`this should get quoted for spaces`,
		`this" should get escaped and quoted`,
		`"this should get escaped and quoted"`,
		`this\ should get quoted`,
		`this' should get quoted`,
	}
	// this is how the strings should be escaped by rsyncEscapeCommand
	stringsShouldMatch := []string{
		`thisshouldnotgetescapedorquoted`,
		`"this should get quoted for spaces"`,
		`"this"" should get escaped and quoted"`,
		`"""this should get escaped and quoted"""`,
		`"this\ should get quoted"`,
		`"this' should get quoted"`,
	}

	escapedStrings := rsyncEscapeCommand(stringsToEscape)

	for key, val := range escapedStrings {
		if val != stringsShouldMatch[key] {
			t.Errorf("%v did not match %v", val, stringsShouldMatch[key])
		}
	}
}
