package unmarshalledmatchers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/format"
	"strings"
)

type ExpandedJsonMatcher struct {
	JSONToMatch      interface{}
	firstFailurePath []interface{}
	DeepMatcher      UnmarshalledDeepMatcher
}

func (matcher *ExpandedJsonMatcher) Match(actual interface{}) (success bool, err error) {
	actualString, expectedString, err := matcher.prettyPrint(actual)
	if err != nil {
		return false, err
	}

	var aval interface{}
	var eval interface{}

	// this is guarded by prettyPrint
	json.Unmarshal([]byte(actualString), &aval)
	json.Unmarshal([]byte(expectedString), &eval)
	var equal bool

	equal, matcher.firstFailurePath = matcher.DeepMatcher.deepEqual(eval, aval)
	return equal, nil
}

func (matcher *ExpandedJsonMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return formattedMessage(format.Message(actualString, "to match JSON of", expectedString), matcher.firstFailurePath)
}

func (matcher *ExpandedJsonMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return formattedMessage(format.Message(actualString, "not to match JSON of", expectedString), matcher.firstFailurePath)
}

func (matcher *ExpandedJsonMatcher) prettyPrint(actual interface{}) (actualFormatted, expectedFormatted string, err error) {
	actualString, ok := toString(actual)
	if !ok {
		return "", "", fmt.Errorf("ExpandedJsonMatcher matcher requires a string, stringer, or []byte.  Got actual:\n%s", format.Object(actual, 1))
	}
	expectedString, ok := toString(matcher.JSONToMatch)
	if !ok {
		return "", "", fmt.Errorf("ExpandedJsonMatcher matcher requires a string, stringer, or []byte.  Got expected:\n%s", format.Object(matcher.JSONToMatch, 1))
	}

	abuf := new(bytes.Buffer)
	ebuf := new(bytes.Buffer)

	if err := json.Indent(abuf, []byte(actualString), "", "  "); err != nil {
		return "", "", fmt.Errorf("Actual '%s' should be valid JSON, but it is not.\nUnderlying error:%s", actualString, err)
	}

	if err := json.Indent(ebuf, []byte(expectedString), "", "  "); err != nil {
		return "", "", fmt.Errorf("Expected '%s' should be valid JSON, but it is not.\nUnderlying error:%s", expectedString, err)
	}

	return abuf.String(), ebuf.String(), nil
}

func formattedMessage(comparisonMessage string, failurePath []interface{}) string {
	var diffMessage string
	if len(failurePath) == 0 {
		diffMessage = ""
	} else {
		diffMessage = fmt.Sprintf("\n\nfirst mismatched key: %s", formattedFailurePath(failurePath))
	}
	return fmt.Sprintf("%s%s", comparisonMessage, diffMessage)
}

func formattedFailurePath(failurePath []interface{}) string {
	formattedPaths := []string{}
	for i := len(failurePath) - 1; i >= 0; i-- {
		switch p := failurePath[i].(type) {
		case int:
			formattedPaths = append(formattedPaths, fmt.Sprintf(`[%d]`, p))
		default:
			if i != len(failurePath)-1 {
				formattedPaths = append(formattedPaths, ".")
			}
			formattedPaths = append(formattedPaths, fmt.Sprintf(`"%s"`, p))
		}
	}
	return strings.Join(formattedPaths, "")
}

func toString(a interface{}) (string, bool) {
	aString, isString := a.(string)
	if isString {
		return aString, true
	}

	aBytes, isBytes := a.([]byte)
	if isBytes {
		return string(aBytes), true
	}

	aStringer, isStringer := a.(fmt.Stringer)
	if isStringer {
		return aStringer.String(), true
	}

	return "", false
}

