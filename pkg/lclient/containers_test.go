package lclient

import (
	"testing"
)

func TestGetContainersByComponentName(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	tests := []struct {
		name      string
		client    *Client
		component string
		wantErr   bool
	}{
		{
			name:      "Case 1: Successfully retrieve one container and have proper component",
			client:    fakeClient,
			component: "node",
			wantErr:   false,
		},
		{
			name:      "Case 2: Successfully retrieve container, but invalid component",
			client:    fakeClient,
			component: "fake-component",
			wantErr:   false,
		},
		{
			name:      "Case 3: Fail to retrieve container",
			client:    fakeErrorClient,
			component: "fake-component",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		container, err := tt.client.GetContainerByComponentName(tt.component)

		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, wanted %v", err, tt.wantErr)
		}

		if container != nil {

		}
	}
}
