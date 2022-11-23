package labels

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/redhat-developer/odo/pkg/version"
)

func Test_getLabels(t *testing.T) {
	type args struct {
		componentName     string
		applicationName   string
		additional        bool
		isPartOfComponent bool
	}
	tests := []struct {
		name string
		args args
		want labels.Set
	}{
		{
			name: "everything filled",
			args: args{
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      false,
			},
			want: labels.Set{
				kubernetesManagedByLabel: "odo",
				kubernetesPartOfLabel:    "applicationame",
				kubernetesInstanceLabel:  "componentname",
				"odo.dev/mode":           "Dev",
			},
		}, {
			name: "everything with additional",
			args: args{
				componentName:   "componentname",
				applicationName: "applicationame",
				additional:      true,
			},
			want: labels.Set{
				kubernetesPartOfLabel:           "applicationame",
				appLabel:                        "applicationame",
				kubernetesManagedByLabel:        "odo",
				kubernetesManagedByVersionLabel: version.VERSION,
				kubernetesInstanceLabel:         "componentname",
				"odo.dev/mode":                  "Dev",
			},
		},
		{
			name: "everything with isPartOfComponent",
			args: args{
				componentName:     "componentname",
				applicationName:   "applicationname",
				isPartOfComponent: true,
			},
			want: labels.Set{
				kubernetesManagedByLabel: "odo",
				kubernetesPartOfLabel:    "applicationname",
				kubernetesInstanceLabel:  "componentname",
				"odo.dev/mode":           "Dev",
				componentLabel:           "componentname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLabels(tt.args.componentName, tt.args.applicationName, ComponentDevMode, tt.args.additional, tt.args.isPartOfComponent)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getLabels() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_sanitizeLabelValue(t *testing.T) {
	type args struct {
		value string
	}
	for _, tt := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid label",
			args: args{value: "a-valid-value"},
			want: "a-valid-value",
		},
		{
			name: "empty value",
			args: args{value: ""},
			want: "",
		},
		{
			name: "blank value",
			args: args{value: "  \t"},
			want: "",
		},
		{
			name: "dot",
			args: args{value: "."},
			want: "dot",
		},
		{
			name: "leading dot - upper",
			args: args{value: ".NET"},
			want: "DOTNET",
		},
		{
			name: "leading dot - lower",
			args: args{value: ".net"},
			want: "dotnet",
		},
		{
			name: "leading dot - mixed",
			args: args{value: ".NeT"},
			want: "dotNeT",
		},
		{
			name: "trailing dot - upper",
			args: args{value: "NET."},
			want: "NETDOT",
		},
		{
			name: "trailing dot - lower",
			args: args{value: "net."},
			want: "netdot",
		},
		{
			name: "trailing dot - mixed",
			args: args{value: "NeT."},
			want: "NeTdot",
		},
		{
			name: "leading and trailing dots",
			args: args{value: ".."},
			want: "dotdot",
		},
		{
			name: "leading and trailing dots",
			args: args{value: ".NET."},
			want: "DOTNETDOT",
		},
		{
			name: "sharp",
			args: args{value: "#"},
			want: "sharp",
		},
		{
			name: "sharpdot",
			args: args{value: "#."},
			want: "sharpdot",
		},
		{
			name: "dotsharp",
			args: args{value: ".#"},
			want: "dotsharp",
		},
		{
			name: "leading sharp - upper",
			args: args{value: "#NET"},
			want: "SHARPNET",
		},
		{
			name: "leading sharp - lower",
			args: args{value: "#net"},
			want: "sharpnet",
		},
		{
			name: "leading sharp - mixed",
			args: args{value: "#NeT"},
			want: "sharpNeT",
		},
		{
			name: "trailing sharp - upper",
			args: args{value: "C#"},
			want: "CSHARP",
		},
		{
			name: "trailing sharp - lower",
			args: args{value: "c#"},
			want: "csharp",
		},
		{
			name: "trailing sharp - mixed",
			args: args{value: "NeT#"},
			want: "NeTsharp",
		},
		{
			name: "leading and trailing sharps",
			args: args{value: "##"},
			want: "sharpsharp",
		},
		{
			name: "single invalid character",
			args: args{value: "?"},
			want: "",
		},
		{
			name: "leading and trailing sharps with valid character in between",
			args: args{value: "#C#"},
			want: "SHARPCSHARP",
		},

		{
			name: "leading non alpha-numeric",
			args: args{value: "-something"},
			want: "something",
		},
		{
			name: "trailing non alpha-numeric",
			args: args{value: "some thing-"},
			want: "some-thing",
		},
		{
			name: "more than 63 characters",
			args: args{value: "Express.js (a.k.a Express), the de facto standard server framework for Node.js"},
			want: "Express.js--a.k.a-Express---the-de-facto-standard-server-framew",
		},
		{
			name: "more than 63 characters starting with space",
			args: args{value: " Express.js (a.k.a Express), the de facto standard server framework for Node.js"},
			want: "Express.js--a.k.a-Express---the-de-facto-standard-server-framew",
		},
		{
			name: "more than 63 characters where truncating it leads to string ending with non-alphanumeric",
			args: args{value: " Express.js (a.k.a Express), the de facto standard server frame-work"},
			want: "Express.js--a.k.a-Express---the-de-facto-standard-server-frame",
		},
		{
			name: "more than 63 characters starting and ending with replacable char should be truncated after replacement",
			args: args{
				value: ".NET (Lorem Ipsum Dolor Sit Amet, consectetur adipiscing elit, " +
					"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.) C#",
			},
			want: "dotNET--Lorem-Ipsum-Dolor-Sit-Amet--consectetur-adipiscing-elit",
		},
		{
			name: "more than 63 characters ending with a lot of invalid characters",
			args: args{value: ".NET" + strings.Repeat(" ", 60)},
			want: "DOTNET",
		},
		{
			name: "more than 63 invalid-only characters",
			args: args{value: strings.Repeat("/[@\\", 90)},
			want: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLabelValue(openshiftRunTimeLabel, tt.args.value)
			if got != tt.want {
				t.Errorf("unexpected value for label. Expected %q, got %q", tt.want, got)
			}
			validationErrs := validation.IsValidLabelValue(got)
			if len(validationErrs) != 0 {
				t.Errorf("expected result %q to be a valid label value, but got the following errors: %s",
					got, strings.Join(validationErrs, "; "))
			}
		})
	}
}
