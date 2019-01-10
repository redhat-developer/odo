package occlient

import "testing"

func TestHasTag(t *testing.T) {
	cases := []struct {
		list     []string
		inputTag string
		expected bool
	}{
		{
			list:     []string{"builder", "php", "hidden"},
			inputTag: "hidden",
			expected: true,
		},
		{
			list:     []string{"builder", "nodejs", "hidden"},
			inputTag: "php",
			expected: false,
		},
	}

	for _, testCase := range cases {
		outcome := hasTag(testCase.list, testCase.inputTag)
		if outcome != testCase.expected {
			t.Errorf("hasTag(%v, %v) returned %v, expected %v",
				testCase.list, testCase.inputTag, outcome, testCase.expected)

		}
	}
}
