package log

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_wrapWarningMessage(t *testing.T) {
	type args struct {
		fullMessage string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty message",
			args: args{
				fullMessage: "",
			},
			want: "",
		},
		{
			name: "single-line message",
			args: args{
				fullMessage: "Lorem Ipsum Dolor Sit Amet",
			},
			want: `==========================
Lorem Ipsum Dolor Sit Amet
==========================`,
		},
		{
			name: "multi-line message",
			args: args{
				fullMessage: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Aenean vel faucibus ex.
Nulla in magna sem.
Vivamus condimentum ultricies erat, in ullamcorper risus tempor nec.
Nam sed risus blandit,
`,
			},
			want: `====================================================================

Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Aenean vel faucibus ex.
Nulla in magna sem.
Vivamus condimentum ultricies erat, in ullamcorper risus tempor nec.
Nam sed risus blandit,

====================================================================`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapWarningMessage(tt.args.fullMessage)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("wrapWarningMessage() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
