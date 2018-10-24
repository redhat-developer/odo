package util

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"testing"
)

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

func TestFilePathConversion(t *testing.T) {

	tests := []struct {
		testName  string
		os        string
		direction string
		path      string
		url       string
	}{
		{
			testName:  "Test %q conversion for %q",
			os:        "windows",
			direction: "url to path",
			path:      "c:\\file\\path\\windows\\test",
			url:       "file:///c:/file/path/windows/test",
		},
		{
			testName:  "Test %s conversion for %s",
			os:        "windows",
			direction: "path to url",
			path:      "c:\\file\\path\\windows\\test",
			url:       "file:///c:/file/path/windows/test",
		},
		{
			testName:  "Test %q conversion for %q",
			os:        "linux",
			direction: "url to path",
			path:      "/c/file/path/windows/test",
			url:       "file:///c/file/path/windows/test",
		},
		{
			testName:  "Test %q conversion for %q",
			os:        "linux",
			direction: "path to url",
			path:      "/c/file/path/windows/test",
			url:       "file:///c/file/path/windows/test",
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf(tt.testName, tt.direction, tt.os)
		t.Log("Running test: ", testName)
		t.Run(testName, func(t *testing.T) {
			if tt.direction == "url to path" {
				url, err := url.Parse(tt.url)
				if err == nil {
					path := ReadFilePath(url, tt.os)
					if path != tt.path {
						t.Errorf(fmt.Sprintf("Expected an url '%s' to be converted to a path '%s", tt.url, tt.path))
					}
				} else {
					t.Errorf(fmt.Sprintf("Error when parsing url '%s'", tt.url))
				}
			} else if tt.direction == "path to url" {
				url := GenFileUrl(tt.path, tt.os)
				if url != tt.url {
					t.Errorf(fmt.Sprintf("Expected a path to be '%s' converted to an url '%s", tt.path, tt.url))
				}
			} else {
				t.Errorf(fmt.Sprintf("Unexpected direction '%s'", tt.direction))
			}
		})
	}
}

func TestParametersAsMap(t *testing.T) {

	tests := []struct {
		testName    string
		sliceInput  []string
		expectedMap map[string]string
	}{
		{
			testName:    "empty slice",
			sliceInput:  []string{},
			expectedMap: map[string]string{},
		},
		{
			testName:   "slice with single element",
			sliceInput: []string{"name=value"},
			expectedMap: map[string]string{
				"name": "value",
			},
		},
		{
			testName:   "slice with multiple elements",
			sliceInput: []string{"name1=value1", "name2=value2", "name3=value3"},
			expectedMap: map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			resultingMap := ConvertKeyValueStringToMap(tt.sliceInput)
			if !reflect.DeepEqual(tt.expectedMap, resultingMap) {
				t.Errorf("Expected %s, got %s", tt.expectedMap, resultingMap)
			}
		})
	}

}

func TestGetDNS1123Name(t *testing.T) {

	tests := []struct {
		testName string
		param    string
		want     string
	}{
		{
			testName: "Case: Test get DNS-1123 name for namespace and version qualified imagestream",
			param:    "myproject/foo:3.5",
			want:     "myproject-foo-3-5",
		},
		{
			testName: "Case: Test get DNS-1123 name for doubly hyphenated string",
			param:    "nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := GetDNS1123Name(tt.param)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}

}

func TestGetRandomName(t *testing.T) {
	type args struct {
		prefix    string
		existList []string
	}
	tests := []struct {
		testName string
		args     args
		// want is regexp if expectConflictResolution is true else its a full name
		want string
	}{
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is not already used",
			args: args{
				prefix: "odo",
				existList: []string{
					"odo-auth",
					"odo-pqrs",
				},
			},
			want: "odo-[a-z]{4}",
		},
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is already used",
			args: args{
				prefix: "nodejs-ex-nodejs",
				existList: []string{
					"nodejs-ex-nodejs-yvrp",
					"nodejs-ex-nodejs-abcd",
				},
			},
			want: "nodejs-ex-nodejs-[a-z]{4}",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := GetRandomName(tt.args.prefix, -1, tt.args.existList, 3)
			if err != nil {
				t.Errorf("failed to generate a random name. Error %v", err)
			}

			r, _ := regexp.Compile(tt.want)
			match := r.MatchString(name)
			if !match {
				t.Errorf("Received name %s which does not match %s", name, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		testName  string
		str       string
		strLength int
		want      string
	}{
		{
			testName:  "Case: Truncate string to greater length",
			str:       "qw",
			strLength: 4,
			want:      "qw",
		},
		{
			testName:  "Case: Truncate string to lesser length",
			str:       "rtyu",
			strLength: 3,
			want:      "rty",
		},
		{
			testName:  "Case: Truncate string to -1 length",
			str:       "Odo",
			strLength: -1,
			want:      "Odo",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			receivedStr := TruncateString(tt.str, tt.strLength)
			if tt.want != receivedStr {
				t.Errorf("Truncated string %s is not same as %s", receivedStr, tt.want)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		testName  string
		strLength int
	}{
		{
			testName:  "Case: Generate random string of length 4",
			strLength: 4,
		},
		{
			testName:  "Case: Generate random string of length 3",
			strLength: 3,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := GenerateRandomString(tt.strLength)
			r, _ := regexp.Compile(fmt.Sprintf("[a-z]{%d}", tt.strLength))
			match := r.MatchString(name)
			if !match {
				t.Errorf("Randomly generated string %s which does not match regexp %s", name, fmt.Sprintf("[a-z]{%d}", tt.strLength))
			}
		})
	}
}

func TestSliceDifference(t *testing.T) {
	tests := []struct {
		testName       string
		slice1         []string
		slice2         []string
		expectedResult []string
	}{
		{
			testName:       "Empty slices",
			slice1:         []string{},
			slice2:         []string{},
			expectedResult: []string{},
		},
		{
			testName:       "Single different slices",
			slice1:         []string{"a"},
			slice2:         []string{"b"},
			expectedResult: []string{"b"},
		},
		{
			testName:       "Single same slices",
			slice1:         []string{"a"},
			slice2:         []string{"a"},
			expectedResult: []string{},
		},
		{
			testName:       "Large slices with matching and non matching items",
			slice1:         []string{"a", "b", "c", "d", "e"},
			slice2:         []string{"e", "a", "u", "1", "d"},
			expectedResult: []string{"u", "1"},
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			result := SliceDifference(tt.slice1, tt.slice2)
			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}
