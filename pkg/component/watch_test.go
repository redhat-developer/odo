package component

import (
	"testing"
)

func TestIsRegExpMatch(t *testing.T) {

	tests := []struct {
		testName   string
		strToMatch string
		regExps    []string
		want       bool
		wantErr    bool
	}{
		{
			testName:   "Test regexp matches",
			strToMatch: "/home/redhat/git-srcs/src/github.com/redhat-developer/nodejs-ex/.git/",
			regExps:    []string{".*\\.git.*", "tests"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test regexp does not match",
			strToMatch: "/home/redhat/git-srcs/src/github.com/redhat-developer/nodejs-ex/gimmt.gimmt/",
			regExps:    []string{".*\\.git.*", "tests"},
			want:       false,
			wantErr:    false,
		},
		{
			testName:   "Test incorrect regexp",
			strToMatch: "a(b",
			regExps:    []string{"a(b"},
			want:       false,
			wantErr:    true,
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			matched, err := isRegExpMatch(tt.strToMatch, tt.regExps)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != matched {
				t.Errorf("Expected %v, got %v", tt.want, matched)
			}
		})
	}

}
