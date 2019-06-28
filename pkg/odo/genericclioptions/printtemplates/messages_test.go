package printtemplates

import (
	"testing"
)

func Test_validatePushMessage(t *testing.T) {
	type args struct {
		action, what string
	}
	tests := []struct {
		name     string
		argl     args
		expected string
	}{
		{
			name: "create route push message",
			argl: args{
				action: "create",
				what:   "route",
			},
			expected: "To create route on the OpenShift Cluster, please use `odo push` \n",
		},
	}

	for _, tt := range tests {
		actual := PushMessage(tt.argl.action, tt.argl.what)
		if actual != tt.expected {
			t.Errorf("Expected : %s Actual %s Mismatch", tt.expected, actual)
		}
	}
}
