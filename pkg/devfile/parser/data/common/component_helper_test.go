package common

import (
	"testing"
)

func TestIsContainer(t *testing.T) {

	tests := []struct {
		name            string
		component       DevfileComponent
		wantIsSupported bool
	}{
		{
			name: "Case 1: Container component",
			component: DevfileComponent{
				Name:      "name",
				Container: &Container{}},
			wantIsSupported: true,
		},
		{
			name: "Case 2: Not a container component",
			component: DevfileComponent{
				Openshift: &Openshift{},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := tt.component.IsContainer()
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsContainer error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}

func TestIsVolume(t *testing.T) {

	tests := []struct {
		name            string
		component       DevfileComponent
		wantIsSupported bool
	}{
		{
			name: "Case 1: Volume component",
			component: DevfileComponent{
				Name: "name",
				Volume: &Volume{
					Size: "size",
				}},
			wantIsSupported: true,
		},
		{
			name: "Case 2: Not a volume component",
			component: DevfileComponent{
				Openshift: &Openshift{},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := tt.component.IsVolume()
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsVolume error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}
