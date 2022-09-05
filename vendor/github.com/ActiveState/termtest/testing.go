// Copyright 2020 ActiveState Software. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file

package termtest

import (
	"fmt"
	"testing"

	expect "github.com/ActiveState/termtest/expect"
	"github.com/ActiveState/termtest/internal/osutils/stacktrace"
)

// TestSendObserveFn is an example for a SendObserver function, it reports any error during Send calls to the supplied testing instance
func TestSendObserveFn(t *testing.T) func(string, int, error) {
	return func(msg string, num int, err error) {
		if err == nil {
			return
		}

		t.Fatalf("Could not send data to terminal\nerror: %v", err)
	}
}

// TestExpectObserveFn an example for a ExpectObserver function, it reports any error occurring durint expect calls to the supplied testing instance
func TestExpectObserveFn(t *testing.T) expect.ExpectObserver {
	return func(matchers []expect.Matcher, ms *expect.MatchState, err error) {
		if err == nil {
			return
		}

		var value string
		var sep string
		for _, matcher := range matchers {
			value += fmt.Sprintf("%s%v", sep, matcher.Criteria())
			sep = ", "
		}

		t.Fatalf(
			"Could not meet expectation: Expectation: '%s'\nError: %v at\n%s\n---\nTerminal snapshot:\n%s\n---\nParsed output:\n%+q\n",
			value, err, stacktrace.Get().String(), ms.TermState.StringBeforeCursor(), ms.Buf.String(),
		)
	}
}
