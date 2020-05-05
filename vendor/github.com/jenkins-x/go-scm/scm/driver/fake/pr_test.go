package fake

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestListChangesPagination(t *testing.T) {
	prNum := 11
	pageTests := []struct {
		prNum     int
		items     int
		page      int
		size      int
		wantFiles []string
	}{
		{prNum, 10, 2, 5, []string{"file6", "file7", "file8", "file9", "file10"}},
		{50, 10, 2, 5, []string{}},
	}

	for i, tt := range pageTests {
		t.Run(fmt.Sprintf("[%d]", i+1), func(rt *testing.T) {
			ctx := context.Background()
			client, data := NewDefault()
			// This stores the data in the "prNum" PR, but the list gets it from
			// the test number.
			data.PullRequestChanges[prNum] = makeChanges(tt.items)

			items, _, err := client.PullRequests.ListChanges(ctx, "test/test", tt.prNum, scm.ListOptions{Page: tt.page, Size: tt.size})
			if err != nil {
				t.Error(err)
				return
			}
			if got := extractChangeFiles(items); !reflect.DeepEqual(got, tt.wantFiles) {
				rt.Errorf("ListChanges() got %#v, want %#v", got, tt.wantFiles)
			}
		})
	}
}

func TestPaginated(t *testing.T) {
	tests := []struct {
		page      int
		size      int
		items     int
		wantStart int
		wantEnd   int
	}{
		{1, 5, 10, 0, 5},
		{2, 5, 10, 5, 10},
		{2, 5, 9, 5, 9},
		{4, 5, 10, 10, 10}, // this results in an empty slice
		{0, 0, 10, 0, 10},  // this is the default 0 value for ListOption
	}

	for _, tt := range tests {
		start, end := paginated(tt.page, tt.size, tt.items)
		if tt.wantStart != start || tt.wantEnd != end {
			t.Fatalf("paginaged(%d, %d, %d) got items[%d:%d], want items[%d:%d]", tt.page, tt.size, tt.items, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}

func makeChanges(n int) []*scm.Change {
	c := []*scm.Change{}
	for i := 1; i <= n; i++ {
		c = append(c, &scm.Change{
			Path: fmt.Sprintf("file%d", i),
		})
	}
	return c
}

func extractChangeFiles(ch []*scm.Change) []string {
	f := []string{}
	for _, c := range ch {
		f = append(f, c.Path)
	}
	return f
}
