package cmd

import "testing"

func Test_validateName(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		testCase string
		name     string
		wantErr  bool
	}{
		{
			testCase: "Test case - 1",
			name:     "app",
			wantErr:  false,
		},
		{
			testCase: "Test case - 2",
			name:     "app123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 3",
			name:     "app-123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 4",
			name:     "app.123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 5",
			name:     "app_123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 6",
			name:     "App",
			wantErr:  true,
		},
		{
			testCase: "Test case - 7",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6",
			wantErr:  false,
		},
		{
			testCase: "Test case - 8",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6p",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Log("Running test", tt.testCase)
		t.Run(tt.testCase, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("Expected error = %v, But got = %v", tt.wantErr, err)
			}
		})
	}
}
