package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMerge(t *testing.T) {
	mergeTests := []struct {
		src  Resources
		dest Resources
		want Resources
	}{
		{
			src:  Resources{"test1": "val1"},
			dest: Resources{},
			want: Resources{"test1": "val1"},
		},
		{
			src:  Resources{"test1": "val1"},
			dest: Resources{"test2": "val2"},
			want: Resources{"test1": "val1", "test2": "val2"},
		},
		{
			src:  Resources{"test1": "val1"},
			dest: Resources{"test1": "val2"},
			want: Resources{"test1": "val1"},
		},
	}

	for _, tt := range mergeTests {
		result := Merge(tt.src, tt.dest)

		if diff := cmp.Diff(tt.want, result); diff != "" {
			t.Fatalf("failed merge: %s\n", diff)
		}
	}

}
