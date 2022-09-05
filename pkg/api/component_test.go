package api

import "testing"

func TestRunningModeList_String(t *testing.T) {
	tests := []struct {
		name string
		o    RunningModeList
		want string
	}{
		{
			name: "only dev",
			o: RunningModeList{
				"dev":    true,
				"deploy": false,
			},
			want: "Dev",
		},
		{
			name: "only deploy",
			o: RunningModeList{
				"dev":    false,
				"deploy": true,
			},
			want: "Deploy",
		},
		{
			name: "both",
			o: RunningModeList{
				"dev":    true,
				"deploy": true,
			},
			want: "Dev, Deploy",
		},
		{
			name: "none",
			o: RunningModeList{
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
