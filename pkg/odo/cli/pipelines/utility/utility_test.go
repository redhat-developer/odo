package utility

import "testing"

func TestAddGitSuffix(t *testing.T) {
	tt := []struct {
		name string
		url  string
		want string
	}{
		{"missing git suffix", "https://github.com/test/org", "https://github.com/test/org.git"},
		{"suffix for empty string", "", ""},
		{"suffix already present", "https://github.com/test/org.git", "https://github.com/test/org.git"},
		{"suffix with a different case", "https://github.com/test/org.GIT", "https://github.com/test/org.GIT"},
	}

	for _, test := range tt {
		t.Run(test.name, func(rt *testing.T) {
			got := AddGitSuffixIfNecessary(test.url)
			if test.want != got {
				rt.Fatalf("URL mismatch: got %s, want %s", got, test.want)
			}
		})
	}
}
