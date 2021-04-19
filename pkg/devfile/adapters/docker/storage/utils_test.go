package storage

import (
	"strings"
	"testing"
)

func TestGenerateVolName(t *testing.T) {

	tests := []struct {
		name        string
		volName     string
		cmpName     string
		wantVolName string
		wantErr     bool
	}{
		{
			name:        "Case 1: Valid volume and component name",
			volName:     "myVol",
			cmpName:     "myCmp",
			wantVolName: "myVol-myCmp",
			wantErr:     false,
		},
		{
			name:        "Case 2: Valid volume name, empty component name",
			volName:     "myVol",
			cmpName:     "",
			wantVolName: "myVol-",
			wantErr:     false,
		},
		{
			name:        "Case 3: Long Valid volume and component name",
			volName:     "myVolmyVolmyVolmyVolmyVolmyVolmyVolmyVolmyVol",
			cmpName:     "myCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmp",
			wantVolName: "myVolmyVolmyVolmyVolmyVolmyVolmyVolmyVolmyVol-",
			wantErr:     false,
		},
		{
			name:        "Case 4: Empty volume name",
			volName:     "",
			cmpName:     "myCmp",
			wantVolName: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedVolName, err := GenerateVolName(tt.volName, tt.cmpName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGenerateVolName error: unexpected error when generating volume name: %v", err)
			} else if !tt.wantErr && !strings.Contains(generatedVolName, tt.wantVolName) {
				t.Errorf("TestGenerateVolName error: generating volume name does not semi match wanted volume name, wanted: %s got: %s", tt.wantVolName, generatedVolName)
			}
		})
	}

}
