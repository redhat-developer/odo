package api

import "testing"

func TestRunningModeList_String(t *testing.T) {
	tests := []struct {
		name string
		o    RunningModes
		want string
	}{
		{
			name: "only dev",
			o: RunningModes{
				"dev":    true,
				"deploy": false,
			},
			want: "Dev",
		},
		{
			name: "only deploy",
			o: RunningModes{
				"dev":    false,
				"deploy": true,
			},
			want: "Deploy",
		},
		{
			name: "both",
			o: RunningModes{
				"dev":    true,
				"deploy": true,
			},
			want: "Dev, Deploy",
		},
		{
			name: "none",
			o: RunningModes{
				"dev":    false,
				"deploy": false,
			},
			want: "None",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.String(); got != tt.want {
				t.Errorf("RunningModeList.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
