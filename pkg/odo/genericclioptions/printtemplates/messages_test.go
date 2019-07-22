package printtemplates

import (
	"testing"
)

func Test_validatePushMessage(t *testing.T) {
	type args struct {
		action, what string
		config  bool
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
				config: false,
			},
			expected: "To create route on the OpenShift Cluster, please use `odo push` \n",
		},
		{
			name: "create config push message",
			argl: args{
				action: "apply",
				what: "config changes",
				config: true,
			},
			expected: "To apply config changes on the OpenShift Cluster, please use `odo push --config` \n",
		},
	}

	for _, tt := range tests {
		actual := PushMessage(tt.argl.action, tt.argl.what, tt.argl.config)
		if actual != tt.expected {
			t.Errorf("Expected : %s Actual %s Mismatch", tt.expected, actual)
		}
	}
}
