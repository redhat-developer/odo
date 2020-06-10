package helper

import (
	"regexp"
	"testing"
)

// AssertErrorMatch returns true if an error matches the required string.
//
// Fails if the error provided does not match provided message, which can be a
// regular expression.

// e.g. AssertErrorMatch(t, "failed to open", err) would fail with a Fatal error
// if the string did not match.
func AssertErrorMatch(t *testing.T, msg string, testErr error) {
	t.Helper()
	if !ErrorMatch(t, msg, testErr) {
		t.Fatalf("failed to match error: '%s' did not match %v", testErr, msg)
	}
}

// ErrorMatch returns true if an error matches the required string.
//
// e.g. ErrorMatch(t, "failed to open", err) would return true if the
// err passed in had a string that matched.
//
// The message can be a regular expression, and if this fails to compile, then
// the test will fail.
func ErrorMatch(t *testing.T, msg string, testErr error) bool {
	t.Helper()
	if msg == "" && testErr == nil {
		return true
	}
	if msg != "" && testErr == nil {
		return false
	}
	match, err := regexp.MatchString(msg, testErr.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}
