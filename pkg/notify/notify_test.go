package notify

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"reflect"
	"testing"
)

func Test_getLatestReleaseTag(t *testing.T) {
	client := github.NewClient(nil)
	release, response, err := client.Repositories.GetLatestRelease(context.Background(), "redhat-developer", "odo")
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		t.Errorf("error getting latest release TagName via API, error: %v", err)
		return
	}

	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "Fetch version from release page using API and compare",
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			releaseTag, err := getLatestReleaseTag()
			if test.success == true && err == nil {
				if !reflect.DeepEqual(fmt.Sprintf("v%s", releaseTag), *release.TagName) {
					t.Errorf("Expected value is %s, got %s", releaseTag, *release.TagName)
				}
			}
		})
	}
}
