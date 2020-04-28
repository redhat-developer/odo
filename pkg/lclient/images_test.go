package lclient

import (
	"testing"
)

func TestPullImage(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()
	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name:    "Verify docker pull image success",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Verify docker pull image failure",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.PullImage("dummyImage")
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}
