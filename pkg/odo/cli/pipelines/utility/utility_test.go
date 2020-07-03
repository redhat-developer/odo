package utility

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddGitSuffix(t *testing.T) {
	addSuffixTests := []struct {
		name string
		url  string
		want string
	}{
		{"missing git suffix", "https://github.com/test/org", "https://github.com/test/org.git"},
		{"suffix for empty string", "", ""},
		{"suffix already present", "https://github.com/test/org.git", "https://github.com/test/org.git"},
		{"suffix with a different case", "https://github.com/test/org.GIT", "https://github.com/test/org.GIT"},
	}

	for _, tt := range addSuffixTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := AddGitSuffixIfNecessary(tt.url)
			if tt.want != got {
				rt.Fatalf("URL mismatch: got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	stringsTests := []struct {
		name   string
		source []string
		want   []string
	}{
		{"no strings", []string{}, []string{}},
		{"no empty strings", []string{"test1", "test2"}, []string{"test1", "test2"}},
		{"mixed strings", []string{"", "test2", ""}, []string{"test2"}},
	}

	for _, tt := range stringsTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := RemoveEmptyStrings(tt.source)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Fatalf("string removal failed:\n%s", diff)
			}
		})
	}
}

func TestMaybeCompletePrefix(t *testing.T) {
	stringsTests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"with dash on end", "testing-", "testing-"},
		{"with no dash on end", "testing", "testing-"},
	}

	for _, tt := range stringsTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := MaybeCompletePrefix(tt.prefix)
			if tt.want != got {
				rt.Fatalf("prefixing failed, got %#v, want %#v", got, tt.want)
			}
		})
	}
}
