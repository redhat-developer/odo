package util

import "testing"

func TestNamespaceOpenShiftObject(t *testing.T) {

	tests := []struct {
		testName        string
		componentName   string
		applicationName string
		want            string
		wantErr         bool
	}{
		{
			testName:        "Test namespacing",
			componentName:   "foo",
			applicationName: "bar",
			want:            "foo-bar",
		},
		{
			testName:        "Blank applicationName with namespacing",
			componentName:   "foo",
			applicationName: "",
			wantErr:         true,
		},
		{
			testName:        "Blank componentName with namespacing",
			componentName:   "",
			applicationName: "bar",
			wantErr:         true,
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := NamespaceOpenShiftObject(tt.componentName, tt.applicationName)

			if tt.wantErr && err == nil {
				t.Errorf("Expected an error, got success")
			} else if tt.wantErr == false && err != nil {
				t.Errorf("Error with namespacing: %s", err)
			}

			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestExtractComponentType(t *testing.T) {

	tests := []struct {
		testName      string
		componentType string
		want          string
		wantErr       bool
	}{
		{
			testName:      "Test namespacing and versioning",
			componentType: "myproject/foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test versioning",
			componentType: "foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test plain component type",
			componentType: "foo",
			want:          "foo",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := ExtractComponentType(tt.componentType)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestParseCreateCmdArgs(t *testing.T) {

	tests := []struct {
		testName string
		args     []string
		want1    string
		want2    string
		want3    string
		want4    string
	}{
		{
			testName: "Case1: Version not specified",
			args: []string{
				"nodejs",
			},
			want1: "nodejs",
			want2: "nodejs",
			want3: "nodejs",
			want4: "latest",
		},
		{
			testName: "Case1: Version not specified",
			args: []string{
				"python:3.5",
			},
			want1: "python:3.5",
			want2: "python",
			want3: "python",
			want4: "3.5",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			got1, got2, got3, got4 := ParseCreateCmdArgs(tt.args)
			if tt.want1 != got1 {
				t.Errorf("Expected imagename to be: %s, got %s", tt.want1, got1)
			}
			if tt.want2 != got2 {
				t.Errorf("Expected component type to be: %s, got %s", tt.want2, got2)
			}
			if tt.want3 != got3 {
				t.Errorf("Expected component name to be: %s, got %s", tt.want3, got3)
			}
			if tt.want4 != got4 {
				t.Errorf("Expected component version to be: %s, got %s", tt.want4, got4)
			}
		})
	}
}
