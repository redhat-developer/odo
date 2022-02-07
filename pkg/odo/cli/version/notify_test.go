package version

/*
import (
	"fmt"
	"github.com/blang/semver"
	"strings"
	"testing"
)

func Test_getLatestReleaseTag(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "parse version string and see if it returns a validated Version",
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			releaseTag, err := getLatestReleaseTag()

			if test.success == true && err == nil {
				Semver, err := semver.Make(strings.TrimPrefix(releaseTag, "v"))
				if err != nil {
					t.Errorf("unable to make semver from the latest release tag: %v", releaseTag)
				}
				t.Log(fmt.Sprintf("getLatestReleaseTag() returned %v\n", Semver))
			}
		})
	}
}
*/
