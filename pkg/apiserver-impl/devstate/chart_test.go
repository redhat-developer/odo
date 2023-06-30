package devstate

import (
	"testing"
)

func TestDevfileState_GetFlowChart(t *testing.T) {
	tests := []struct {
		name    string
		state   func() DevfileState
		want    string
		wantErr bool
	}{
		{
			name: "with initial devfile",
			state: func() DevfileState {
				return NewDevfileState()
			},
			want: `graph TB
containers["containers"]
start["start"]
sync-all-containers["Sync All Sources"]
start -->|"dev"| containers
containers -->|"container running"| sync-all-containers
`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.GetFlowChart()
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.GetFlowChart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DevfileState.GetFlowChart() = %v, want %v", got, tt.want)
			}
		})
	}
}
