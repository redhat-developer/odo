package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	corev1 "k8s.io/api/core/v1"
)

// TODO(feloy) Move tests to devfile library

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
			name, err := dfutil.NamespaceOpenShiftObject(tt.componentName, tt.applicationName)

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
			name := dfutil.ExtractComponentType(tt.componentType)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
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
		{
			testName: "Case 9: Test get DNS-1123 should remove invalid chars with sufix and prefix",
			param:    "54myproject/$foo@@:3.5",
			want:     "myproject-foo-3-5",
		},
		{
			testName: "Case 9: Test get DNS-1123 should add x as a prefix for all numerics",
			param:    "54453443",
			want:     "x54453443",
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
			name, err := dfutil.GetRandomName(tt.args.prefix, -1, tt.args.existList, 3)
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
		appendStr bool
		appendArr []string
		want      string
	}{
		{
			testName:  "Case: Truncate string to greater length",
			str:       "qw",
			strLength: 4,
			appendStr: false,
			want:      "qw",
		},
		{
			testName:  "Case: Truncate string to lesser length",
			str:       "rtyu",
			strLength: 3,
			appendStr: false,
			want:      "rty",
		},
		{
			testName:  "Case: Truncate string to -1 length",
			str:       "Odo",
			strLength: -1,
			appendStr: false,
			want:      "Odo",
		},
		{
			testName:  "Case: Trunicate string with 3 dots appended",
			str:       "rtyu",
			strLength: 3,
			appendStr: true,
			appendArr: []string{"..."},
			want:      "rty...",
		},
		{
			testName:  "Case: Appends multiple if multiple args are provided",
			str:       "rtyu",
			strLength: 3,
			appendStr: true,
			appendArr: []string{".", ".", "."},
			want:      "rty...",
		},
		{
			testName:  "Case: Does not append if length is lesser than max length",
			str:       "qw",
			strLength: 4,
			appendStr: true,
			appendArr: []string{"..."},
			want:      "qw",
		},
		{
			testName:  "Case: Does not append if maxlength is -1",
			str:       "rtyu",
			strLength: -1,
			appendStr: true,
			appendArr: []string{"..."},
			want:      "rtyu",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			var receivedStr string
			if tt.appendStr {
				receivedStr = TruncateString(tt.str, tt.strLength, tt.appendArr...)
			} else {
				receivedStr = TruncateString(tt.str, tt.strLength)
			}
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
			name := dfutil.GenerateRandomString(tt.strLength)
			r, _ := regexp.Compile(fmt.Sprintf("[a-z]{%d}", tt.strLength))
			match := r.MatchString(name)
			if !match {
				t.Errorf("Randomly generated string %s which does not match regexp %s", name, fmt.Sprintf("[a-z]{%d}", tt.strLength))
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
			result, err := dfutil.GetAbsPath(tt.path)
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
			got, err := dfutil.GetHostWithPort(tt.inputURL)
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
		return fmt.Errorf("error while creating file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("error while writing to file: %w", err)
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
			filesToCreate:    []string{DotGitIgnoreFile},
			rulesOnGitIgnore: "",
			rulesOnOdoIgnore: "",
			wantRules:        []string{".git"},
			wantErr:          false,
		},
		{
			name:             "test case 3: no odoignore but gitignore exists with rules",
			directoryName:    testDir,
			filesToCreate:    []string{DotGitIgnoreFile},
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
			filesToCreate:    []string{DotGitIgnoreFile, ".odoignore"},
			rulesOnGitIgnore: "/tests",
			rulesOnOdoIgnore: "*.json\n\n/openshift/**/*.js",
			wantRules:        []string{".git", "*.json", "/openshift/**/*.js"},
			wantErr:          false,
		},
		{
			name:             "test case 7: no odoignore but gitignore exists with rules and comments",
			directoryName:    testDir,
			filesToCreate:    []string{DotGitIgnoreFile},
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
			if fileName == DotGitIgnoreFile {
				err = MakeFileWithContent(testDir, fileName, tt.rulesOnGitIgnore)
			} else if fileName == ".odoignore" {
				err = MakeFileWithContent(testDir, fileName, tt.rulesOnOdoIgnore)
			}
			if err != nil {
				t.Fatal(err)
			}
		}

		gotRules, err := dfutil.GetIgnoreRulesFromDirectory(testDir)

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
			resultExps := dfutil.GetAbsGlobExps(tt.directoryName, tt.inputRelativeGlobExps)
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
			actual := dfutil.GetSortedKeys(tt.input)
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
			actual := dfutil.GetSplitValuesFromStr(tt.input)
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
			ports, err := dfutil.GetContainerPortsFromStrings(tt.ports)
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
			matched, err := dfutil.IsGlobExpMatch(tt.strToMatch, tt.globExps)

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
			output, err := dfutil.RemoveRelativePathFromFiles(tt.args.input, tt.args.path)
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
			got, err := dfutil.HTTPGetFreePort()
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
			remoteFiles := dfutil.GetRemoteFilesMarkedForDeletion(tt.files, tt.remotePath)
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
			request := dfutil.HTTPRequestParams{
				URL: tt.url,
			}
			got, err := dfutil.HTTPGetRequest(request, 0)

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
			filterChanged, filterDeleted := dfutil.FilterIgnores(tt.changedFiles, tt.deletedFiles, tt.ignoredFiles)

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
			params := dfutil.DownloadParams{
				Request: dfutil.HTTPRequestParams{
					URL: tt.url,
				},
				Filepath: tt.filepath,
			}
			err := dfutil.DownloadFile(params)
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
			err := dfutil.ValidateK8sResourceName(tt.key, tt.value)
			got := err == nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestUnzip(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		pathToUnzip   string
		expectedFiles []string
		expectedError string
	}{
		{
			name:          "Case 1: Invalid source zip",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/invalid.zip",
			pathToUnzip:   "",
			expectedFiles: []string{},
			expectedError: "open ../../tests/examples/source/devfiles/nodejs-zip/invalid.zip:",
		},
		{
			name:          "Case 2: Valid source zip, no pathToUnzip",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "",
			expectedFiles: []string{"package.json", "package-lock.json", "app", "app/app.js", DotGitIgnoreFile, "LICENSE", "README.md"},
			expectedError: "",
		},
		{
			name:          "Case 3: Valid source zip with pathToUnzip",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "app",
			expectedFiles: []string{"app.js"},
			expectedError: "",
		},
		{
			name:          "Case 4: Valid source zip with pathToUnzip - trailing /",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "app/",
			expectedFiles: []string{"app.js"},
			expectedError: "",
		},
		{
			name:          "Case 5: Valid source zip with pathToUnzip - leading /",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "app/",
			expectedFiles: []string{"app.js"},
			expectedError: "",
		},
		{
			name:          "Case 6: Valid source zip with pathToUnzip - leading and trailing /",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "/app/",
			expectedFiles: []string{"app.js"},
			expectedError: "",
		},
		{
			name:          "Case 7: Valid source zip with pathToUnzip - pattern",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "p*",
			expectedFiles: []string{"package.json", "package-lock.json"},
			expectedError: "",
		},
		{
			name:          "Case 8: Valid source zip with pathToUnzip - pattern and extension",
			src:           "../../tests/examples/source/devfiles/nodejs-zip/master.zip",
			pathToUnzip:   "*.json",
			expectedFiles: []string{"package.json", "package-lock.json"},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "unzip")
			if err != nil {
				t.Errorf("Error creating temp dir: %s", err)
			}
			defer os.RemoveAll(dir)
			t.Logf(dir)

			_, err = Unzip(filepath.FromSlash(tt.src), dir, tt.pathToUnzip)
			if err != nil {
				tt.expectedError = strings.ReplaceAll(tt.expectedError, "/", string(filepath.Separator))
				if !strings.HasPrefix(err.Error(), tt.expectedError) {
					t.Errorf("Got err: '%s'\n expected err: '%s'", err.Error(), tt.expectedError)
				}
			} else {
				for _, file := range tt.expectedFiles {
					if _, err := os.Stat(filepath.Join(dir, file)); os.IsNotExist(err) {
						t.Errorf("Expected file %s does not exist in directory after unzipping", file)
					}
				}
			}
		})
	}
}

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
			expectedError: "folder %s doesn't contain the devfile used",
		},
		{
			name:          "Case 4: Folder contains a hidden file which is not the devfile",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{".file1.yaml"},
			dirToCreate:   []string{},
			expectedError: "folder %s doesn't contain the devfile used",
		},
		{
			name:          "Case 5: Folder contains devfile.yaml and more files",
			devfilePath:   "devfile.yaml",
			filesToCreate: []string{"devfile.yaml", "file1.yaml", "file2.yaml"},
			dirToCreate:   []string{},
			expectedError: "folder %s is not empty. It can only contain the devfile used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "valid-project")
			if err != nil {
				t.Errorf("Error creating temp dir: %s", err)
			}
			defer os.RemoveAll(tmpDir)

			for _, f := range tt.filesToCreate {
				file := filepath.Join(tmpDir, f)
				if _, e := os.Create(file); e != nil {
					t.Errorf("Error creating file %s. Err: %s", file, e)
				}
			}

			for _, d := range tt.dirToCreate {
				dir := filepath.Join(tmpDir, d)
				if e := os.Mkdir(dir, os.FileMode(0644)); e != nil {
					t.Errorf("Error creating dir %s. Err: %s", dir, e)
				}
			}

			err = IsValidProjectDir(tmpDir, tt.devfilePath)
			expectedError := tt.expectedError
			if expectedError != "" {
				expectedError = fmt.Sprintf(expectedError, tmpDir)
			}

			if err != nil && !reflect.DeepEqual(err.Error(), expectedError) {
				t.Errorf("Got err: %s, expected err %s", err.Error(), expectedError)
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
			data, err := DownloadFileInMemory(dfutil.HTTPRequestParams{URL: tt.url})
			if tt.url != "invalid" && err != nil {
				t.Errorf("Failed to download file with error %s", err)
			}

			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("Got: %v, want: %v", data, tt.want)
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
		{
			name:    "Case 4: Invalid URL - Host contains reserved character",
			url:     "http://??##",
			wantErr: true,
		},
		{
			name:    "Case 5: Invalid URL - Scheme contains reserved character",
			url:     "$$$,,://www.example.com",
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
			err := dfutil.ValidateFile(tt.filePath)
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
			err = dfutil.CopyFile(tt.srcPath, tt.dstPath, info)
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
			got := dfutil.PathEqual(tt.firstPath, tt.secondPath)
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

// FileType custom type to indicate type of file
type FileType int

const (
	// RegularFile enum to represent regular file
	RegularFile FileType = 0
	// Directory enum to represent directory
	Directory FileType = 1
)

// FileProperties to contain meta-data of a file like, file/folder name, file/folder parent dir, file type and desired file modification type
type FileProperties struct {
	FilePath string
	FileType FileType
}

func folderCheck(originalFolderMode os.FileMode, newFolderInfo os.FileInfo, path string) error {
	if originalFolderMode.String() != newFolderInfo.Mode().String() {
		return fmt.Errorf("folder %s created with wrong permission", path)
	}
	return nil
}

func fileCheck(fs filesystem.Filesystem, originalFile filesystem.File, newFilePath string, newFileInfo os.FileInfo) error {
	originalFileData, err := fs.ReadFile(originalFile.Name())
	if err != nil {
		return err
	}

	createdFileData, err := fs.ReadFile(newFilePath)
	if err != nil {
		return err
	}

	// check the written data
	if string(createdFileData) == string(originalFileData) {
		originalInfo, err := fs.Stat(originalFile.Name())
		if err != nil {
			return err
		}

		// check the file permission
		if newFileInfo.Mode() != originalInfo.Mode() {
			return fmt.Errorf("file %s created with wrong permission", newFilePath)
		}
	} else {
		return fmt.Errorf("file %s created with wrong data", newFilePath)
	}
	return nil
}

func setupFileTest(fs filesystem.Filesystem, sourceName string, filePaths []FileProperties) (map[string]filesystem.File, map[string]os.FileMode, error) {

	fileMap := make(map[string]filesystem.File)
	folderMap := make(map[string]os.FileMode)

	for i, path := range filePaths {

		if path.FileType == RegularFile {
			file, err := fs.Create(path.FilePath)
			if err != nil {
				return nil, nil, err
			}
			_, err = file.Write([]byte("some text" + string(rune(i))))
			if err != nil {
				return nil, nil, err
			}

			name, err := filepath.Rel(sourceName, file.Name())
			if err != nil {
				return nil, nil, err
			}
			fileMap[name] = file
		} else if path.FileType == Directory {
			permission := os.ModeDir + 0755
			err := fs.MkdirAll(path.FilePath, permission)
			if err != nil {
				return nil, nil, err
			}

			name, err := filepath.Rel(sourceName, path.FilePath)
			if err != nil {
				return nil, nil, err
			}
			folderMap[name] = permission
		}
	}

	return fileMap, folderMap, nil
}

func TestCopyFileWithFS(t *testing.T) {
	fileName := "blah.js"

	fs := filesystem.NewFakeFs()

	sourceName := filepath.Join(os.TempDir(), "source")
	destinationDirName := filepath.Join(os.TempDir(), "destination")

	filePaths := []FileProperties{
		{
			FilePath: sourceName,
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, fileName),
			FileType: RegularFile,
		},
		{
			FilePath: destinationDirName,
			FileType: Directory,
		},
	}

	printError := func(err error) {
		t.Errorf("some error occured while procession file/folder: %v", err)
	}

	type args struct {
		src string
		dst string
		fs  filesystem.Filesystem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: normal file exists",
			args: args{
				src: filepath.Join(sourceName, fileName),
				dst: filepath.Join(destinationDirName, fileName),
				fs:  fs,
			},
		},
		{
			name: "case 2: file doesn't exist",
			args: args{
				src: filepath.Join(sourceName, fileName) + "blah",
				dst: filepath.Join(destinationDirName, fileName),
				fs:  fs,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMap, _, err := setupFileTest(fs, sourceName, filePaths)
			if err != nil {
				t.Errorf("error while setting up test: %v", err)
			}

			err = copyFileWithFs(tt.args.src, tt.args.dst, tt.args.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			files, err := fs.ReadDir(destinationDirName)
			if err != nil {
				t.Errorf("error occured while reading directory %s: %v", destinationDirName, err)
			}

			found := false
			for _, file := range files {

				relPath, err := filepath.Rel(destinationDirName, filepath.Join(destinationDirName, fileName))
				if err != nil {
					printError(err)
					break
				}

				if originalFile, ok := fileMap[relPath]; ok {
					err := fileCheck(fs, originalFile, filepath.Join(destinationDirName, file.Name()), file)
					if err != nil {
						break
					}
					found = true
				} else {
					t.Errorf("extra file %s created", file.Name())
				}
			}

			if !found && !tt.wantErr {
				t.Errorf("%s not created in directory %s", fileName, destinationDirName)
			}

			_ = fs.RemoveAll(tt.args.src)
			_ = fs.RemoveAll(tt.args.dst)
		})
	}
}

func TestCopyDirWithFS(t *testing.T) {

	fs := filesystem.NewFakeFs()

	sourceName := filepath.Join(os.TempDir(), "source")
	destinationName := filepath.Join(os.TempDir(), "destination")

	filePaths := []FileProperties{
		{
			FilePath: sourceName,
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "blah.js"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "some-other-blah.js"),
			FileType: RegularFile,
		},
	}

	printError := func(err error) {
		t.Errorf("some error occured while procession file/folder: %v", err)
	}

	type args struct {
		src string
		dst string
		fs  filesystem.Filesystem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: folder with a nested file and folder",
			args: args{
				src: sourceName,
				dst: destinationName,
				fs:  fs,
			},
		},
		{
			name: "case 2: source folder doesn't exist",
			args: args{
				src: sourceName + "/extra",
				dst: destinationName,
				fs:  fs,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMap, folderMap, err := setupFileTest(fs, sourceName, filePaths)
			if err != nil {
				t.Errorf("error while setting up test: %v", err)
			}

			err = copyDirWithFS(tt.args.src, tt.args.dst, tt.args.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				return
			}

			folderCount := 0
			fileCount := 0

			err = fs.Walk(destinationName, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(destinationName, path)
				if err != nil {
					printError(err)
					return err
				}

				if info.IsDir() {

					// check the permission on the folder
					if originalFolderMode, ok := folderMap[relPath]; ok {
						err := folderCheck(originalFolderMode, info, path)
						if err != nil {
							printError(err)
							return err
						}
						folderCount++
					} else {
						t.Errorf("extra folder %s created", path)
					}
				} else {
					if file, ok := fileMap[relPath]; ok {
						err := fileCheck(fs, file, path, info)
						if err != nil {
							printError(err)
							return err
						}
						fileCount++
					} else {
						t.Errorf("extra file %s created", path)
					}
				}
				return nil
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if folderCount != 2 {
				t.Errorf("some folder were not created")
			}

			if fileCount != 2 {
				t.Errorf("some files were not created")
			}

			_ = fs.RemoveAll(tt.args.src)
			_ = fs.RemoveAll(tt.args.dst)
		})
	}
}

func TestCleanDir(t *testing.T) {
	fs := filesystem.NewFakeFs()

	sourceName := filepath.Join(os.TempDir(), "source")

	filePaths := []FileProperties{
		{
			FilePath: sourceName,
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "devfile.yaml"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "preference.yaml"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "blah.js"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "some-other-blah.js"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "another-blah.js"),
			FileType: RegularFile,
		},
	}

	type args struct {
		originalPath     string
		leaveBehindFiles map[string]bool
		fs               filesystem.Filesystem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: leave behind two files",
			args: args{
				originalPath: sourceName,
				fs:           fs,
				leaveBehindFiles: map[string]bool{
					"devfile.yaml":    true,
					"preference.yaml": true,
				},
			},
		},
		{
			name: "case 2: source doesn't exist",
			args: args{
				originalPath: sourceName + "blah",
				fs:           fs,
				leaveBehindFiles: map[string]bool{
					"devfile.yaml":    true,
					"preference.yaml": true,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := setupFileTest(fs, sourceName, filePaths)
			if err != nil {
				t.Errorf("error while setting up test: %v", err)
			}

			err = cleanDir(tt.args.originalPath, tt.args.leaveBehindFiles, tt.args.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			files, err := fs.ReadDir(sourceName)
			if err != nil {
				t.Errorf("error occured while reading directory %s: %v", sourceName, err)
			}

			found := 0
			for _, file := range files {
				if _, ok := tt.args.leaveBehindFiles[file.Name()]; !ok {
					t.Errorf("file %s isn't cleaned up", file.Name())
				} else {
					found++
				}
			}

			if found != 2 {
				t.Errorf("some extra file were deleted")
			}

			_ = fs.RemoveAll(tt.args.originalPath)
		})
	}
}

func TestGitSubDir(t *testing.T) {

	fs := filesystem.NewFakeFs()

	sourceName := filepath.Join(os.TempDir(), "source")
	destinationName := filepath.Join(os.TempDir(), "destination")

	filePaths := []FileProperties{
		{
			FilePath: sourceName,
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "blah.js"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "some-other-blah.js"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "test"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "test", "test.java"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "java"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "java", "main.java"),
			FileType: RegularFile,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "resources"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "resources", "layout"),
			FileType: Directory,
		},
		{
			FilePath: filepath.Join(sourceName, "main", "src", "resources", "index.html"),
			FileType: RegularFile,
		},
	}

	type args struct {
		destinationPath string
		srcPath         string
		subDir          string
		fs              filesystem.Filesystem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: normal sub dir exist",
			args: args{
				srcPath:         sourceName,
				destinationPath: destinationName,
				subDir:          filepath.Join("main", "src"),
				fs:              fs,
			},
		},
		{
			name: "case 2: sub dir doesn't exist",
			args: args{
				srcPath:         sourceName,
				destinationPath: destinationName,
				subDir:          filepath.Join("main", "blah"),
				fs:              fs,
			},
			wantErr: true,
		},
		{
			name: "case 3: src doesn't exist",
			args: args{
				srcPath:         sourceName + "blah",
				destinationPath: destinationName,
				subDir:          filepath.Join("main", "src"),
				fs:              fs,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := setupFileTest(fs, sourceName, filePaths)
			if err != nil {
				t.Errorf("error while setting up test: %v", err)
			}

			err = gitSubDir(tt.args.srcPath, tt.args.destinationPath, tt.args.subDir, tt.args.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitSubDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				return
			}

			pathsToValidate := map[string]bool{
				filepath.Join(tt.args.destinationPath, "java"):                    true,
				filepath.Join(tt.args.destinationPath, "java", "main.java"):       true,
				filepath.Join(tt.args.destinationPath, "resources"):               true,
				filepath.Join(tt.args.destinationPath, "resources", "layout"):     true,
				filepath.Join(tt.args.destinationPath, "resources", "index.html"): true,
			}

			pathsNotToBePresent := map[string]bool{
				filepath.Join(tt.args.destinationPath, "src"):  true,
				filepath.Join(tt.args.destinationPath, "main"): true,
			}

			found := 0
			notToBeFound := 0
			err = fs.Walk(tt.args.destinationPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if ok := pathsToValidate[path]; ok {
					found++
				}

				if ok := pathsNotToBePresent[path]; ok {
					notToBeFound++
				}
				return nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if found != 5 {
				t.Errorf("all files were not copied")
			}

			if notToBeFound != 0 {
				t.Errorf("extra files were created")
			}

			_, err = os.Stat(tt.args.srcPath)
			if !os.IsNotExist(err) {
				t.Errorf("src path was not deleted")
			}

			_ = fs.RemoveAll(tt.args.srcPath)
			_ = fs.RemoveAll(tt.args.destinationPath)
		})
	}
}

func TestGetCommandStringFromEnvs(t *testing.T) {
	type args struct {
		envVars []v1alpha2.EnvVar
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1: three envs given",
			args: args{
				envVars: []v1alpha2.EnvVar{
					{
						Name:  "foo",
						Value: "bar",
					},
					{
						Name:  "JAVA_HOME",
						Value: "/home/user/java",
					},
					{
						Name:  "GOPATH",
						Value: "/home/user/go",
					},
				},
			},
			want: "export foo=\"bar\" JAVA_HOME=\"/home/user/java\" GOPATH=\"/home/user/go\"",
		},
		{
			name: "case 2: no envs given",
			args: args{
				envVars: []v1alpha2.EnvVar{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCommandStringFromEnvs(tt.args.envVars); got != tt.want {
				t.Errorf("GetCommandStringFromEnvs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGitOriginPath(t *testing.T) {
	tempGitDirWithOrigin, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	defer os.RemoveAll(tempGitDirWithOrigin)

	repoWithOrigin, err := git.PlainInit(tempGitDirWithOrigin, true)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = repoWithOrigin.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:redhat-developer/odo.git"},
	})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	tempGitDirWithoutOrigin, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	defer os.RemoveAll(tempGitDirWithoutOrigin)

	repoWithoutOrigin, err := git.PlainInit(tempGitDirWithoutOrigin, true)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = repoWithoutOrigin.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{"git@github.com:redhat-developer/odo.git"},
	})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1: remote named origin exists",
			args: args{
				path: tempGitDirWithOrigin,
			},
			want: "git@github.com:redhat-developer/odo.git",
		},
		{
			name: "case 2: remote named origin doesn't exists",
			args: args{
				path: tempGitDirWithoutOrigin,
			},
			want: "",
		},
		{
			name: "case 3: not a git repo",
			args: args{
				path: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGitOriginPath(tt.args.path); got != tt.want {
				t.Errorf("GetGitOriginPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertLabelsToSelector(t *testing.T) {
	cases := []struct {
		labels map[string]string
		want   string
	}{
		{
			labels: map[string]string{
				"app":                                  "app",
				"app.kubernetes.io/managed-by":         "odo",
				"app.kubernetes.io/managed-by-version": "v2.1",
			},
			want: "app=app,app.kubernetes.io/managed-by=odo,app.kubernetes.io/managed-by-version=v2.1",
		},
		{
			labels: map[string]string{
				"app":                                  "app",
				"app.kubernetes.io/managed-by":         "!odo",
				"app.kubernetes.io/managed-by-version": "4.8",
			},
			want: "app=app,app.kubernetes.io/managed-by!=odo,app.kubernetes.io/managed-by-version=4.8",
		},
		{
			labels: map[string]string{
				"app.kubernetes.io/managed-by": "odo",
			},
			want: "app.kubernetes.io/managed-by=odo",
		},
		{
			labels: map[string]string{
				"app.kubernetes.io/managed-by": "!odo",
			},
			want: "app.kubernetes.io/managed-by!=odo",
		},
	}

	for _, tt := range cases {
		got := ConvertLabelsToSelector(tt.labels)
		if got != tt.want {
			t.Errorf("got: %q\nwant:%q", got, tt.want)
		}
	}
}

func TestNamespaceKubernetesObjectWithTrim(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "case 1: hyphenated name is less than 63 characters",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			want:    "nodejs-app",
			wantErr: false,
		},
		{
			name: "case 2: hyphenated name is more than 63 characters",
			args: args{
				componentName:   "veryveryveryveryveryLongComponentName",
				applicationName: "veryveryveryveryveryveryLongAppName",
			},
			want:    "veryveryveryveryveryLongCompone-veryveryveryveryveryveryLongApp",
			wantErr: false,
		},
		{
			name: "case 3: hyphenated name is equal to 63 characters",
			args: args{
				componentName:   "veryveryveryveryLongComponentGo",
				applicationName: "veryveryveryveryLongAppNameInGo",
			},
			want:    "veryveryveryveryLongComponentGo-veryveryveryveryLongAppNameInGo",
			wantErr: false,
		},
		{
			name: "case 4: component name is more than 63 characters",
			args: args{
				componentName:   "123456789012345678901234567890123456789012345678901234567890ComponentName",
				applicationName: "app",
			},
			want:    "12345678901234567890123456789012345678901234567890123456789-app",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NamespaceKubernetesObjectWithTrim(tt.args.componentName, tt.args.applicationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NamespaceKubernetesObjectWithTrim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NamespaceKubernetesObjectWithTrim() got = %v, want %v", got, tt.want)
			}

			if len(got) > 63 {
				t.Errorf("got = %s should be less than or equal to 63 characters", got)
			}
		})
	}
}

func TestSafeGetBool(t *testing.T) {

	tests := []struct {
		name string
		arg  *bool
		want bool
	}{
		{
			name: "case 1: nil pointer",
			arg:  nil,
			want: false,
		},
		{
			name: "case 2: true",
			arg:  GetBoolPtr(true),
			want: true,
		},
		{
			name: "case 3: false",
			arg:  GetBoolPtr(false),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeGetBool(tt.arg); got != tt.want {
				t.Errorf("SafeGetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisplayLog(t *testing.T) {
	const compName = "my-comp"
	type args struct {
		input    []string
		numLines int
	}
	for _, tt := range []struct {
		name    string
		wantErr bool
		want    []string

		args
	}{
		{
			name: "numberOfLastLines==-1",
			args: args{
				input:    []string{"a", "b", "c"},
				numLines: -1,
			},
			want: []string{"a\n", "b\n", "c\n"},
		},
		{
			name: "numberOfLastLines greater than total number of lines read",
			args: args{
				input:    []string{"one-line"},
				numLines: 10,
			},
			want: []string{"one-line\n"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			for _, s := range tt.input {
				if _, err := b.WriteString(s + "\n"); err != nil {
					t.Errorf(" failed to write input data %q", s)
					return
				}
			}
			var w bytes.Buffer
			err := DisplayLog(false, io.NopCloser(&b), &w, compName, tt.numLines)

			if tt.wantErr != (err != nil) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
				return
			}

			//Read w
			reader := bufio.NewReader(&w)
			var lines []string
			var line string
			for {
				line, err = reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Errorf("unexpected err while reading data: %v", err)
					return
				}
				lines = append(lines, line)
			}
			if !reflect.DeepEqual(lines, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, lines)
			}
		})
	}
}
