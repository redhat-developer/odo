package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

func TestParseComponentImageName(t *testing.T) {

	tests := []struct {
		testName string
		args     string
		want1    string
		want2    string
		want3    string
		want4    string
	}{
		{
			testName: "Case1: Version not specified",
			args:     "nodejs",
			want1:    "nodejs",
			want2:    "nodejs",
			want3:    "nodejs",
			want4:    "latest",
		},
		{
			testName: "Case1: Version not specified",
			args:     "python:3.5",
			want1:    "python:3.5",
			want2:    "python",
			want3:    "python",
			want4:    "3.5",
		},
	}

	// Test that it "joins"

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			got1, got2, got3, got4 := ParseComponentImageName(tt.args)
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
				url := GenFileURL(tt.path, tt.os)
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
			testName: "Case 1: Test get DNS-1123 name for namespace and version qualified imagestream",
			param:    "myproject/foo:3.5",
			want:     "myproject-foo-3-5",
		},
		{
			testName: "Case 2: Test get DNS-1123 name for doubly hyphenated string",
			param:    "nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 3: Test get DNS-1123 name for string starting with underscore",
			param:    "_nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 4: Test get DNS-1123 name for string ending with underscore",
			param:    "nodejs--myproject-foo-3.5_",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 5: Test get DNS-1123 name for string having multiple non alpha-numeric chars as prefix",
			param:    "_#*nodejs--myproject-foo-3.5",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 6: Test get DNS-1123 name for string having multiple non alpha-numeric chars as suffix",
			param:    "nodejs--myproject-foo-3.5=_@",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 7: Test get DNS-1123 name for string having with multiple non alpha-numeric chars as prefix and suffix",
			param:    " _#*nodejs--myproject-foo-3.5=_@ ",
			want:     "nodejs-myproject-foo-3-5",
		},
		{
			testName: "Case 8: Test get DNS-1123 should remove invalid chars",
			param:    "myproject/$foo@@:3.5",
			want:     "myproject-foo-3-5",
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

func TestGetAbsPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		absPath string
		wantErr bool
	}{
		{
			name:    "Case 1: Valid abs path resolution of `~`",
			path:    "~",
			wantErr: false,
		},
		{
			name:    "Case 2: Valid abs path resolution of `.`",
			path:    ".",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			switch tt.path {
			case "~":
				if len(customHomeDir) > 0 {
					tt.absPath = customHomeDir
				} else {
					usr, err := user.Current()
					if err != nil {
						t.Errorf("Failed to get absolute path corresponding to `~`. Error %v", err)
						return
					}
					tt.absPath = usr.HomeDir
				}

			case ".":
				absPath, err := os.Getwd()
				if err != nil {
					t.Errorf("Failed to get absolute path corresponding to `.`. Error %v", err)
					return
				}
				tt.absPath = absPath
			}
			result, err := GetAbsPath(tt.path)
			if result != tt.absPath {
				t.Errorf("Expected %v, got %v", tt.absPath, result)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("Expected error: %v got error %v", tt.wantErr, err)
			}
		})
	}
}

func TestCheckPathExists(t *testing.T) {
	dir, err := ioutil.TempFile("", "")
	defer os.RemoveAll(dir.Name())
	if err != nil {
		return
	}
	tests := []struct {
		fileName string
		wantBool bool
	}{
		{
			fileName: dir.Name(),
			wantBool: true,
		},
		{
			fileName: dir.Name() + "-blah",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		exists := CheckPathExists(tt.fileName)
		if tt.wantBool != exists {
			t.Errorf("the expected value of TestCheckPathExists function is different : %v, got: %v", tt.wantBool, exists)
		}
	}
}

func TestGetHostWithPort(t *testing.T) {

	tests := []struct {
		inputURL string
		want     string
		wantErr  bool
	}{
		{
			inputURL: "https://example.com",
			want:     "example.com:443",
			wantErr:  false,
		},
		{
			inputURL: "https://example.com:8443",
			want:     "example.com:8443",
			wantErr:  false,
		},
		{
			inputURL: "http://example.com",
			want:     "example.com:80",
			wantErr:  false,
		},
		{
			inputURL: "notexisting://example.com",
			want:     "",
			wantErr:  true,
		},
		{
			inputURL: "http://127.0.0.1",
			want:     "127.0.0.1:80",
			wantErr:  false,
		},
		{
			inputURL: "example.com:1234",
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Testing inputURL: %s", tt.inputURL), func(t *testing.T) {
			got, err := GetHostWithPort(tt.inputURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostWithPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostWithPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

// MakeFileWithContent creates file with a given name in the given directory and writes the content to it
// dir is the name of the directory
// fileName is the name of the file to be created
// content is the string to be written to the file
func MakeFileWithContent(dir string, fileName string, content string) error {
	file, err := os.Create(dir + string(os.PathSeparator) + fileName)
	if err != nil {
		return errors.Wrapf(err, "error while creating file")
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return errors.Wrapf(err, "error while writing to file")
	}
	return nil
}

// RemoveContentsFromDir removes content from the given directory
// dir is the name of the directory
func RemoveContentsFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestGetIgnoreRulesFromDirectory(t *testing.T) {
	testDir, err := ioutil.TempDir(os.TempDir(), "odo-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)
	tests := []struct {
		name             string
		directoryName    string
		filesToCreate    []string
		rulesOnGitIgnore string
		rulesOnOdoIgnore string
		wantRules        []string
		wantErr          bool
	}{
		{
			name:             "test case 1: no odoignore and no gitignore",
			directoryName:    testDir,
			filesToCreate:    []string{""},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 2: no odoignore but gitignore exists with no rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 3: no odoignore but gitignore exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "*.js\n\n/openshift/**/*.json\n/tests",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json", "/tests"},
			wantErr:          false,
		},
		{
			name:             "test case 4: odoignore exists with no rules",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 5: odoignore exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "*.json\n\n/openshift/**/*.js",
			wantRules:        []string{".git", "*.json", "/openshift/**/*.js"},
			wantErr:          false,
		},
		{
			name:             "test case 6: odoignore and gitignore both exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore", ".odoignore"},
			rulesOnGitIgnore: "/tests",
			rulesOnOdoIgnore: "*.json\n\n/openshift/**/*.js",
			wantRules:        []string{".git", "*.json", "/openshift/**/*.js"},
			wantErr:          false,
		},
		{
			name:             "test case 7: no odoignore but gitignore exists with rules and comments",
			directoryName:    testDir,
			filesToCreate:    []string{".gitignore"},
			rulesOnGitIgnore: "*.js\n\n/openshift/**/*.json\n\n\n#/tests",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json"},
			wantErr:          false,
		},
		{
			name:             "test case 8: odoignore exists exists with rules and comments",
			directoryName:    testDir,
			filesToCreate:    []string{".odoignore"},
			rulesOnOdoIgnore: "*.js\n\n\n/openshift/**/*.json\n\n\n#/tests\n/bin",
			rulesOnGitIgnore: "",
			wantRules:        []string{".git", "*.js", "/openshift/**/*.json", "/bin"},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		for _, fileName := range tt.filesToCreate {
			var err error
			if fileName == ".gitignore" {
				err = MakeFileWithContent(testDir, fileName, tt.rulesOnGitIgnore)
			} else if fileName == ".odoignore" {
				err = MakeFileWithContent(testDir, fileName, tt.rulesOnOdoIgnore)
			}
			if err != nil {
				t.Fatal(err)
			}
		}

		gotRules, err := GetIgnoreRulesFromDirectory(testDir)

		if err == nil && !tt.wantErr {
			if !reflect.DeepEqual(gotRules, tt.wantRules) {
				t.Errorf("the expected value of rules are different, excepted: %v, got: %v", tt.wantRules, gotRules)
			}
		} else if err == nil && tt.wantErr {
			t.Error("error was expected, but no error was returned")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
		}
		err = RemoveContentsFromDir(testDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetAbsGlobExps(t *testing.T) {
	tests := []struct {
		testName              string
		directoryName         string
		inputRelativeGlobExps []string
		expectedGlobExps      []string
	}{
		{
			testName:      "test case 1: with a filename",
			directoryName: "/home/redhat/nodejs-ex/",
			inputRelativeGlobExps: []string{
				"example.txt",
			},
			expectedGlobExps: []string{
				"/home/redhat/nodejs-ex/example.txt",
			},
		},
		{
			testName:      "test case 2: with a folder name",
			directoryName: "/home/redhat/nodejs-ex/",
			inputRelativeGlobExps: []string{
				"example/",
			},
			expectedGlobExps: []string{
				"/home/redhat/nodejs-ex/example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			resultExps := GetAbsGlobExps(tt.directoryName, tt.inputRelativeGlobExps)
			if runtime.GOOS == "windows" {
				for index, element := range resultExps {
					resultExps[index] = filepath.ToSlash(element)
				}
			}

			if !reflect.DeepEqual(resultExps, tt.expectedGlobExps) {
				t.Errorf("expected %v, got %v", tt.expectedGlobExps, resultExps)
			}
		})
	}
}

func TestGetSortedKeys(t *testing.T) {
	tests := []struct {
		testName string
		input    map[string]string
		expected []string
	}{
		{
			testName: "default",
			input:    map[string]string{"a": "av", "c": "cv", "b": "bv"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual := GetSortedKeys(tt.input)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, actual)
			}
		})
	}
}

func TestGetSplitValuesFromStr(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected []string
	}{
		{
			testName: "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			testName: "Single value",
			input:    "s1",
			expected: []string{"s1"},
		},
		{
			testName: "Multiple values",
			input:    "s1, s2, s3 ",
			expected: []string{"s1", "s2", "s3"},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual := GetSplitValuesFromStr(tt.input)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, actual)
			}
		})
	}
}

func TestGetContainerPortsFromStrings(t *testing.T) {
	tests := []struct {
		name           string
		ports          []string
		containerPorts []corev1.ContainerPort
		wantErr        bool
	}{
		{
			name:  "with normal port values and normal protocol values in lowercase",
			ports: []string{"8080/tcp", "9090/udp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:  "with normal port values and normal protocol values in mixed case",
			ports: []string{"8080/TcP", "9090/uDp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:  "with normal port values and with one protocol value not mentioned",
			ports: []string{"8080", "9090/Udp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:    "with normal port values and with one invalid protocol value",
			ports:   []string{"8080/blah", "9090/Udp"},
			wantErr: true,
		},
		{
			name:    "with invalid port values and normal protocol",
			ports:   []string{"ads/Tcp", "9090/Udp"},
			wantErr: true,
		},
		{
			name:    "with invalid port values and one missing protocol value",
			ports:   []string{"ads", "9090/Udp"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := GetContainerPortsFromStrings(tt.ports)
			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(tt.containerPorts, ports) {
					t.Errorf("the ports are not matching, expected %#v, got %#v", tt.containerPorts, ports)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestIsGlobExpMatch(t *testing.T) {

	tests := []struct {
		testName   string
		strToMatch string
		globExps   []string
		want       bool
		wantErr    bool
	}{
		{
			testName:   "Test glob matches",
			strToMatch: "/home/redhat/nodejs-ex/.git",
			globExps:   []string{"/home/redhat/nodejs-ex/.git", "/home/redhat/nodejs-ex/tests/"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test glob does not match",
			strToMatch: "/home/redhat/nodejs-ex/gimmt.gimmt",
			globExps:   []string{"/home/redhat/nodejs-ex/.git/", "/home/redhat/nodejs-ex/tests/"},
			want:       false,
			wantErr:    false,
		},
		{
			testName:   "Test glob match files",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/*.json", "/home/redhat/nodejs-ex/tests/"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test '**' glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/openshift/**/*.json"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test '!' in glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/!*.json", "/home/redhat/nodejs-ex/tests/"},
			want:       false,
			wantErr:    false,
		},
		{
			testName:   "Test [ in glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/["},
			want:       false,
			wantErr:    true,
		},
		{
			testName:   "Test '#' comment glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"#/home/redhat/nodejs-ex/openshift/**/*.json"},
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			matched, err := IsGlobExpMatch(tt.strToMatch, tt.globExps)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != matched {
				t.Errorf("expected %v, got %v", tt.want, matched)
			}
		})
	}
}

func TestRemoveDuplicate(t *testing.T) {
	type args struct {
		input  []string
		output []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Case 1 - Remove duplicates",
			args: args{
				input:  []string{"bar", "bar"},
				output: []string{"bar"},
			},
		},
		{
			name: "Case 2 - Remove duplicates, none in array",
			args: args{
				input:  []string{"bar", "foo"},
				output: []string{"foo", "bar"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Run function RemoveDuplicate
			output := RemoveDuplicates(tt.args.input)

			// sort the strings
			sort.Strings(output)
			sort.Strings(tt.args.output)

			if !(reflect.DeepEqual(output, tt.args.output)) {
				t.Errorf("expected %v, got %v", tt.args.output, output)
			}

		})
	}
}

func TestRemoveRelativePathFromFiles(t *testing.T) {
	type args struct {
		path   string
		input  []string
		output []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1 - Remove the relative path from a list of files",
			args: args{
				path:   "/foo/bar",
				input:  []string{"/foo/bar/1", "/foo/bar/2", "/foo/bar/3/", "/foo/bar/4/5/foo/bar"},
				output: []string{"1", "2", "3", "4/5/foo/bar"},
			},
			wantErr: false,
		},
		{
			name: "Case 2 - Fail on purpose with an invalid path",
			args: args{
				path:   `..`,
				input:  []string{"foo", "bar"},
				output: []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Run function RemoveRelativePathFromFiles
			output, err := RemoveRelativePathFromFiles(tt.args.input, tt.args.path)
			if runtime.GOOS == "windows" {
				for index, element := range output {
					output[index] = filepath.ToSlash(element)
				}
			}

			// Check for error
			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !(reflect.DeepEqual(output, tt.args.output)) {
				t.Errorf("expected %v, got %v", tt.args.output, output)
			}

		})
	}
}

func TestHTTPGetFreePort(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "case 1: get a valid free port",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTTPGetFreePort()
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPGetFreePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			addressLook := "localhost:" + strconv.Itoa(got)
			listener, err := net.Listen("tcp", addressLook)
			if err != nil {
				t.Errorf("expected a free port, but listening failed cause: %v", err)
			} else {
				_ = listener.Close()
			}
		})
	}
}

func TestGetRemoteFilesMarkedForDeletion(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		remotePath string
		want       []string
	}{
		{
			name:       "case 1: no files",
			files:      []string{},
			remotePath: "/projects",
			want:       nil,
		},
		{
			name:       "case 2: one file",
			files:      []string{"abc.txt"},
			remotePath: "/projects",
			want:       []string{"/projects/abc.txt"},
		},
		{
			name:       "case 3: multiple files",
			files:      []string{"abc.txt", "def.txt", "hello.txt"},
			remotePath: "/projects",
			want:       []string{"/projects/abc.txt", "/projects/def.txt", "/projects/hello.txt"},
		},
		{
			name:       "case 4: remote path multiple folders",
			files:      []string{"abc.txt", "def.txt", "hello.txt"},
			remotePath: "/test/folder",
			want:       []string{"/test/folder/abc.txt", "/test/folder/def.txt", "/test/folder/hello.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteFiles := GetRemoteFilesMarkedForDeletion(tt.files, tt.remotePath)
			if !reflect.DeepEqual(tt.want, remoteFiles) {
				t.Errorf("Expected %s, got %s", tt.want, remoteFiles)
			}
		})
	}
}

func TestHTTPGetRequest(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name string
		url  string
		want []byte
	}{
		{
			name: "Case 1: Input url is valid",
			url:  server.URL,
			// Want(Expected) result is "OK"
			// According to Unicode table: O == 79, K == 75
			want: []byte{79, 75},
		},
		{
			name: "Case 2: Input url is invalid",
			url:  "invalid",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := HTTPRequestParams{
				URL: tt.url,
			}
			got, err := HTTPGetRequest(request)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestFilterIgnores(t *testing.T) {
	tests := []struct {
		name             string
		changedFiles     []string
		deletedFiles     []string
		ignoredFiles     []string
		wantChangedFiles []string
		wantDeletedFiles []string
	}{
		{
			name:             "Case 1: No ignored files",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{},
			wantChangedFiles: []string{"hello.txt", "test.abc"},
			wantDeletedFiles: []string{"one.txt", "two.txt"},
		},
		{
			name:             "Case 2: One ignored file",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{"hello.txt"},
			wantChangedFiles: []string{"test.abc"},
			wantDeletedFiles: []string{"one.txt", "two.txt"},
		},
		{
			name:             "Case 3: Multiple ignored file",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{"hello.txt", "two.txt"},
			wantChangedFiles: []string{"test.abc"},
			wantDeletedFiles: []string{"one.txt"},
		},
		{
			name:             "Case 4: No changed or deleted files",
			changedFiles:     []string{""},
			deletedFiles:     []string{""},
			ignoredFiles:     []string{"hello.txt", "two.txt"},
			wantChangedFiles: []string{""},
			wantDeletedFiles: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterChanged, filterDeleted := FilterIgnores(tt.changedFiles, tt.deletedFiles, tt.ignoredFiles)

			if !reflect.DeepEqual(tt.wantChangedFiles, filterChanged) {
				t.Errorf("Expected %s, got %s", tt.wantChangedFiles, filterChanged)
			}

			if !reflect.DeepEqual(tt.wantDeletedFiles, filterDeleted) {
				t.Errorf("Expected %s, got %s", tt.wantDeletedFiles, filterDeleted)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name     string
		url      string
		filepath string
		want     []byte
		wantErr  bool
	}{
		{
			name:     "Case 1: Input url is valid",
			url:      server.URL,
			filepath: "./test.yaml",
			// Want(Expected) result is "OK"
			// According to Unicode table: O == 79, K == 75
			want:    []byte{79, 75},
			wantErr: false,
		},
		{
			name:     "Case 2: Input url is invalid",
			url:      "invalid",
			filepath: "./test.yaml",
			want:     []byte{},
			wantErr:  true,
		},
		{
			name:     "Case 3: Input url is an empty string",
			url:      "",
			filepath: "./test.yaml",
			want:     []byte{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			params := DownloadParams{
				Request: HTTPRequestParams{
					URL: tt.url,
				},
				Filepath: tt.filepath,
			}
			err := DownloadFile(params)
			if err != nil {
				gotErr = true
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Error("Failed to get expected error")
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Failed to download file with error %s", err)
				}

				got, err := ioutil.ReadFile(tt.filepath)
				if err != nil {
					t.Errorf("Failed to read file with error %s", err)
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Got: %v, want: %v", got, tt.want)
				}

				// Clean up the file that downloaded in this test case
				err = os.Remove(tt.filepath)
				if err != nil {
					t.Errorf("Failed to delete file with error %s", err)
				}
			}
		})
	}
}

func TestValidateK8sResourceName(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  bool
	}{
		{
			name:  "Case 1: Resource name is valid",
			key:   "component name",
			value: "good-name",
			want:  true,
		},
		{
			name:  "Case 2: Resource name contains unsupported character",
			key:   "component name",
			value: "BAD@name",
			want:  false,
		},
		{
			name:  "Case 3: Resource name contains all numeric values",
			key:   "component name",
			value: "12345",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateK8sResourceName(tt.key, tt.value)
			got := err == nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestConvertGitSSHRemotetoHTTPS(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedResult string
	}{
		{
			name:           "Case 1: Git ssh url is valid",
			url:            "git@github.com:che-samples/web-nodejs-sample.git",
			expectedResult: "https://github.com/che-samples/web-nodejs-sample.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertGitSSHRemoteToHTTPS(tt.url)
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Got %s, want %s", result, tt.expectedResult)
			}
		})
	}
}

// TODO: FIX THIS
/*
func TestUnzip(t *testing.T) {
	tests := []struct {
		name          string
		zipURL        string
		zipDst        string
		expectedFiles []string
	}{
		{
			name:          "Case 1: Valid zip ",
			zipURL:        "https://github.com/che-samples/web-nodejs-sample/archive/master.zip",
			zipDst:        "master.zip",
			expectedFiles: []string{"package.json", "package-lock.json", "app", ".gitignore", "LICENSE", "README.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "unzip")
			if err != nil {
				t.Errorf("Error creating temp dir: %s", err)
			}
			//defer os.RemoveAll(dir)
			t.Logf(dir)

			tt.zipDst = filepath.Join(dir, tt.zipDst)
			err = DownloadFile(tt.zipURL, tt.zipDst)
			if err != nil {
				t.Errorf("Error downloading zip: %s", err)
			}
			_, err = Unzip(tt.zipDst, dir)
			if err != nil {
				t.Errorf("Error unzipping: %s", err)
			}

			for _, file := range tt.expectedFiles {
				if _, err := os.Stat(filepath.Join(dir, file)); os.IsNotExist(err) {
					t.Errorf("Expected file %s does not exist in directory after unzipping", file)
				}
			}
		})
	}
}
*/

func TestIsValidProjectDir(t *testing.T) {
	tests := []struct {
		name          string
		devfilePath   string
		filesToCreate []string
		dirToCreate   []string
		expectedError string
	}{
		{
			name:          "Case 1: Empty Folder",
			devfilePath:   "",
			filesToCreate: []string{},
			dirToCreate:   []string{},
			expectedError: "",
		},
		{
			name:          "Case 2: Folder contains devfile.yaml",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{"devfile.yaml"},
			dirToCreate:   []string{},
			expectedError: "",
		},
		{
			name:          "Case 3: Folder contains a file which is not the devfile",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{"file1.yaml"},
			dirToCreate:   []string{},
			expectedError: "Folder contains one element and it's not the devfile used.",
		},
		{
			name:          "Case 4: Folder contains a hidden file which is not the devfile",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{".file1.yaml"},
			dirToCreate:   []string{},
			expectedError: "Folder contains one element and it's not the devfile used.",
		},
		{
			name:          "Case 5: Folder contains devfile.yaml and more files",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{"devfile.yaml", "file1.yaml", "file2.yaml"},
			dirToCreate:   []string{},
			expectedError: "Folder is not empty. It can only contain the devfile used.",
		},
		{
			name:          "Case 6: Folder contains a directory",
			devfilePath:   "",
			filesToCreate: []string{},
			dirToCreate:   []string{"dir"},
			expectedError: "Folder is not empty. It contains a subfolder.",
		},
		{
			name:          "Case 7: Folder contains a hidden directory",
			devfilePath:   "",
			filesToCreate: []string{},
			dirToCreate:   []string{".dir"},
			expectedError: "Folder is not empty. It contains a subfolder.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "valid-project")
			if err != nil {
				t.Errorf("Error creating temp dir: %s", err)
			}
			defer os.RemoveAll(tmpDir)

			for _, file := range tt.filesToCreate {
				file := filepath.Join(tmpDir, file)
				_, err := os.Create(file)
				if err != nil {
					t.Errorf("Error creating file %s. Err: %s", file, err)
				}
			}

			for _, dir := range tt.dirToCreate {
				dir := filepath.Join(tmpDir, dir)
				err := os.Mkdir(dir, os.FileMode(644))
				if err != nil {
					t.Errorf("Error creating dir %s. Err: %s", dir, err)
				}
			}

			err = IsValidProjectDir(tmpDir, tt.devfilePath)
			if err != nil && !reflect.DeepEqual(err.Error(), tt.expectedError) {
				t.Errorf("Got err: %s, expected err %s", err.Error(), tt.expectedError)
			}
		})
	}
}

func TestDownloadFileInMemory(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name string
		url  string
		want []byte
	}{
		{
			name: "Case 1: Input url is valid",
			url:  server.URL,
			want: []byte{79, 75},
		},
		{
			name: "Case 2: Input url is invalid",
			url:  "invalid",
			want: []byte(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := DownloadFileInMemory(tt.url)

			if tt.url != "invalid" && err != nil {
				t.Errorf("Failed to download file with error %s", err)
			}

			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("Got: %v, want: %v", data, tt.want)
			}
		})
	}
}

func TestLoadFileIntoMemory(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name          string
		url           string
		contains      []byte
		expectedError string
	}{
		{
			name:          "Case 1: Input url is valid",
			url:           server.URL,
			contains:      []byte{79, 75},
			expectedError: "",
		},
		{
			name:          "Case 2: Input url is invalid",
			url:           "invalid",
			contains:      []byte(nil),
			expectedError: "invalid url:",
		},
		{
			name:          "Case 3: Input http:// url doesnt exist",
			url:           "http://test.it.doesnt/exist/",
			contains:      []byte(nil),
			expectedError: "unable to download url",
		},
		{
			name:          "Case 4: Input file:// url doesnt exist",
			url:           "file://./notexists.txt",
			contains:      []byte(nil),
			expectedError: "unable to read file",
		},
		{
			name:          "Case 5: Input file://./util.go exists",
			url:           "file://./util.go",
			contains:      []byte("Load a file into memory ("),
			expectedError: "",
		},
		{
			name:          "Case 5: Input url is empty",
			url:           "",
			contains:      []byte(nil),
			expectedError: "invalid url:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadFileIntoMemory(tt.url)

			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Got err: %s, expected err %s", err.Error(), tt.expectedError)
			}

			if tt.expectedError == "" && !bytes.Contains(data, tt.contains) {
				t.Errorf("Got: %v, should contain: %v", data, tt.contains)
			}
		})
	}
}

/*
func TestGetGitHubZipURL(t *testing.T) {
	startPoint := "1.0.0"
	branch := "my-branch"
	tests := []struct {
		name          string
		location      string
		branch        string
		startPoint    string
		expectedError string
	}{
		{
			name:          "Case 1: Invalid http request",
			location:      "http://github.com/che-samples/web-nodejs-sample/archive/master",
			expectedError: "Invalid GitHub URL. Please use https://",
		},
		{
			name:          "Case 2: Invalid owner",
			location:      "https://github.com//web-nodejs-sample/archive/master",
			expectedError: "Invalid GitHub URL: owner cannot be empty. Expecting 'https://github.com/<owner>/<repo>'",
		},
		{
			name:          "Case 3: Invalid repo",
			location:      "https://github.com/che-samples//archive/master",
			expectedError: "Invalid GitHub URL: repo cannot be empty. Expecting 'https://github.com/<owner>/<repo>'",
		},
		{
			name:          "Case 4: Invalid HTTPS Github URL with tag and commit",
			location:      "https://github.com/che-samples/web-nodejs-sample.git",
			branch:        branch,
			startPoint:    startPoint,
			expectedError: fmt.Sprintf("Branch %s and StartPoint %s specified as project reference, please only specify one", branch, startPoint),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetGitHubZipURL(tt.location, tt.branch, tt.startPoint)
			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Got %s,\n want %s", err.Error(), tt.expectedError)
				}
			}
		})
	}
}
*/

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Case 1: Valid URL",
			url:     "http://www.example.com/",
			wantErr: false,
		},
		{
			name:    "Case 2: Invalid URL - No host",
			url:     "http://",
			wantErr: true,
		},
		{
			name:    "Case 3: Invalid URL - No scheme",
			url:     "://www.example.com/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			got := ValidateURL(tt.url)
			if got != nil {
				gotErr = true
			}

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	// Create temp dir and temp file
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		t.Errorf("Failed to create temp file: %s, error: %v", tempFile.Name(), err)
	}
	defer tempFile.Close()

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Case 1: Valid file path",
			filePath: tempFile.Name(),
			wantErr:  false,
		},
		{
			name:     "Case 2: Invalid file path",
			filePath: "!@#",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			err := ValidateFile(tt.filePath)
			if err != nil {
				gotErr = true
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got error: %t, want error: %t", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp dir
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}

	// Create temp file under temp dir as source file
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		t.Errorf("Failed to create temp file: %s, error: %v", tempFile.Name(), err)
	}
	defer tempFile.Close()

	srcPath := tempFile.Name()
	fakePath := "!@#/**"
	dstPath := filepath.Join(tempDir, "dstFile")
	info, _ := os.Stat(srcPath)

	tests := []struct {
		name    string
		srcPath string
		dstPath string
		wantErr bool
	}{
		{
			name:    "Case 1: Copy successfully",
			srcPath: srcPath,
			dstPath: dstPath,
			wantErr: false,
		},
		{
			name:    "Case 2: Invalid source path",
			srcPath: fakePath,
			dstPath: dstPath,
			wantErr: true,
		},
		{
			name:    "Case 3: Invalid destination path",
			srcPath: srcPath,
			dstPath: fakePath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			err = CopyFile(tt.srcPath, tt.dstPath, info)
			if err != nil {
				gotErr = true
			}

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got error: %t, want error: %t", gotErr, tt.wantErr)
			}
		})
	}
}

func TestPathEqual(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Can't get absolute path of current working directory with error: %v", err)
	}
	fileAbsPath := filepath.Join(currentDir, "file")
	fileRelPath := filepath.Join(".", "file")

	tests := []struct {
		name       string
		firstPath  string
		secondPath string
		want       bool
	}{
		{
			name:       "Case 1: Two paths (two absolute paths) are equal",
			firstPath:  fileAbsPath,
			secondPath: fileAbsPath,
			want:       true,
		},
		{
			name:       "Case 2: Two paths (one absolute path, one relative path) are equal",
			firstPath:  fileAbsPath,
			secondPath: fileRelPath,
			want:       true,
		},
		{
			name:       "Case 3: Two paths are not equal",
			firstPath:  fileAbsPath,
			secondPath: filepath.Join(fileAbsPath, "file"),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathEqual(tt.firstPath, tt.secondPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSliceContainsString(t *testing.T) {
	tests := []struct {
		name      string
		stringVal string
		slice     []string
		wantVal   bool
	}{
		{
			name:      "Case 1: string in valid slice",
			stringVal: "string",
			slice:     []string{"string", "string2"},
			wantVal:   true,
		},
		{
			name:      "Case 2: string not in valid slice",
			stringVal: "string3",
			slice:     []string{"string", "string2"},
			wantVal:   false,
		},
		{
			name:      "Case 3: string not in empty slice",
			stringVal: "string",
			slice:     []string{},
			wantVal:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal := sliceContainsString(tt.stringVal, tt.slice)

			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("Got %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestDownloadInMemory(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Case 1: valid URL",
			url:  "https://github.com/openshift/odo/blob/master/tests/examples/source/devfiles/nodejs/devfile.yaml",
			want: true,
		},
		{
			name: "Case 2: invalid URL",
			url:  "https://this/is/not/a/valid/url",
			want: false,
		},
		{
			name: "Case 3: empty URL",
			url:  "",
			want: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			_, err := DownloadFileInMemory(tt.url)

			got := err == nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestValidateDockerfile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Case 1: valid Dockerfile",
			path: filepath.Join("tests", "examples", "source", "dockerfiles", "Dockerfile"),
			want: true,
		},
		{
			name: "Case 2: valid Dockerfile with comment",
			path: filepath.Join("tests", "examples", "source", "dockerfiles", "DockerfileWithComment"),
			want: true,
		},
		{
			name: "Case 3: valid Dockerfile with whitespace",
			path: filepath.Join("tests", "examples", "source", "dockerfiles", "DockerfileWithWhitespace"),
			want: true,
		},
		{
			name: "Case 4: invalid Dockerfile with missing FROM",
			path: filepath.Join("tests", "examples", "source", "dockerfiles", "DockerfileInvalid"),
			want: false,
		},
		{
			name: "Case 5: invalid Dockerfile with entry before FROM",
			path: filepath.Join("tests", "examples", "source", "dockerfiles", "DockerfileInvalidFROM"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get path for this file (util_test)
			_, filename, _, _ := runtime.Caller(0)
			// Read the file using a path relative to this file
			content, err := ioutil.ReadFile(filepath.Join(filename, "..", "..", "..", tt.path))
			if err != nil {
				t.Error("Error when reading the dockerfile: ", err)
			}

			err = ValidateDockerfile(content)

			got := err == nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestValidateTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{
			name: "Case 1: Valid tag ",
			tag:  "image-registry.openshift-image-registry.svc:5000/default/my-nodejs:1.0",
			want: true,
		},
		{
			name: "Case 2: Invalid tag with trailing period",
			tag:  "image-registry.openshift-image-registry.svc:5000./default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 3: Invalid tag with trailing dash",
			tag:  "image-registry.openshift-image-registry.svc:5000-/default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 4: Invalid tag with trailing underscore",
			tag:  "image-registry.openshift-image-registry.svc:5000_/default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 5: Invalid tag with trailing colon",
			tag:  "image-registry.openshift-image-registry.svc:5000:/default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 6: Invalid tag with invalid characters",
			tag:  "imag|||\\e-registry.openshift&^%-image-registry.svc:5000/default!/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 7: Missing registry",
			tag:  "/default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 8: Missing namespace",
			tag:  "image-registry.openshift-image-registry.svc:5000//my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 9: Missing image",
			tag:  "image-registry.openshift-image-registry.svc:5000/default/",
			want: false,
		},
		{
			name: "Case 10: Too many /'s",
			tag:  "image-registry.openshift/image-registry.svc:5000:/default/my-nodejs:1.0",
			want: false,
		},
		{
			name: "Case 11: Too few /'s",
			tag:  "image-registry.openshift-image-registry.svc:5000:/default-my-nodejs:1.0",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTag(tt.tag)

			got := err == nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}
