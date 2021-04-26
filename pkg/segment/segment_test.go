package segment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/openshift/odo/pkg/testingutil/filesystem"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/version"
)

type segmentResponse struct {
	Batch []struct {
		UserId    string `json:"userId"`
		MessageId string `json:"messageId"`
		Traits    struct {
			OS string `json:"os"`
		} `json:"traits"`
		Properties struct {
			Error     string `json:"error"`
			ErrorType string `json:"error-type"`
			Success   bool   `json:"success"`
			Version   string `json:"version"`
		} `json:"properties"`
		Type string `json:"type"`
	} `json:"batch"`
	MessageID string `json:"messageId"`
}

func mockServer() (chan []byte, *httptest.Server) {
	done := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bin, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			return
		}
		done <- bin
	}))
	return done, server
}

func TestClientUploadWithoutConsent(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)
	falseValue := false

	cfg := &preference.PreferenceInfo{
		Preference: preference.Preference{
			OdoSettings: preference.OdoSettings{
				ConsentTelemetry: &falseValue,
			},
		},
	}
	c, err := newCustomClient(cfg, createConfigDir(t), server.URL)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	testError := errors.New("error occurred")
	uploadData := TelemetryData{
		Event: "odo preference view",
		Properties: TelmetryProperties{
			Error:     SetError(testError),
			ErrorType: ErrorType(testError),
			Success:   false,
			Tty:       RunningInTerminal(),
			Version:   version.VERSION,
		},
	}
	// run a command, odo preference view
	if err = c.Upload(uploadData); err != nil {
		t.Error(err)
	}

	select {
	case x := <-body:
		t.Errorf("server should not receive data: %q", x)
	default:
	}
}

func TestClientUploadWithConsent(t *testing.T) {
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	trueValue := true

	cfg := &preference.PreferenceInfo{
		Preference: preference.Preference{
			OdoSettings: preference.OdoSettings{
				ConsentTelemetry: &trueValue,
			},
		},
	}
	tests := []struct {
		cmd      string
		testName string
		err      error
		success  bool
		errType  string
		version  string
	}{
		{
			testName: "command ran successfully",
			err:      nil,
			success:  true,
			errType:  "",
			version:  version.VERSION,
		},
		{
			testName: "command failed",
			err:      errors.New("some error occurred"),
			success:  false,
			errType:  "*errors.errorString",
			version:  version.VERSION,
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			c, err := newCustomClient(cfg, createConfigDir(t), server.URL)
			if err != nil {
				t.Error(err)
			}
			defer c.Close()
			uploadData := TelemetryData{
				Event: "odo create",
				Properties: TelmetryProperties{
					Duration:  time.Second.Milliseconds(),
					Error:     SetError(tt.err),
					ErrorType: ErrorType(tt.err),
					Success:   tt.err == nil,
					Tty:       RunningInTerminal(),
					Version:   version.VERSION,
				},
			}
			// upload the data to Segment
			if err = c.Upload(uploadData); err != nil {
				t.Error(err)
			}
			// Note: Do not use `default` inside select with for loop, it lands in panic
			for i := 0; i < 2; i++ {
				select {
				case x := <-body:
					s := segmentResponse{}
					if err = json.Unmarshal(x, &s); err != nil {
						t.Error(err)
					}
					fmt.Println(string(x), s)
					//Response returns 2 Batches in response - 1) identify - user's system information in case it is new,
					//and 2) track - information about the fired command
					//This checks if both the responses were received
					if s.Batch[0].Type == "identify" {
						if s.Batch[0].Traits.OS != runtime.GOOS {
							t.Error("OS does not match")
						}
					} else if s.Batch[0].Type == "track" {
						if !tt.success {
							if s.Batch[0].Properties.Error != tt.err.Error() {
								t.Error("Error does not match")
							}
						} else {
							if s.Batch[0].Properties.Error != "" {
								t.Error("Error does not match")
							}
						}
						if s.Batch[0].Properties.Success != tt.success {
							t.Error("Success does not match")
						}
						if s.Batch[0].Properties.ErrorType != tt.errType {
							t.Error("Error Type does not match")
						}
						if !strings.Contains(s.Batch[0].Properties.Version, version.VERSION) {
							t.Error("Odo version does not match")
						}
					} else {
						t.Errorf("Missing Identify or Track information. Available info: %v", s.Batch[0].Type)
					}
				}
			}
		})
	}
}

func TestSetError(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error(err.Error())
	}
	unixPath := "/home/xyz/.odo/preference.yaml"
	windowsPath := "C:\\User\\XYZ\\preference.yaml"

	tests := []struct {
		name   string
		err    error
		hasPII bool
	}{
		{
			name:   "no PII information",
			err:    errors.New("this is an error string"),
			hasPII: false,
		},
		{
			name:   "username",
			err:    fmt.Errorf("cannot create component name with %s", user.Username),
			hasPII: true,
		},
		{
			name:   "filepath-unix",
			err:    fmt.Errorf("cannot find the preference file at %s", unixPath),
			hasPII: true,
		},
		{
			name:   "filepath-windows",
			err:    fmt.Errorf("cannot find the preference file at %s", windowsPath),
			hasPII: true,
		},
	}

	for _, tt := range tests {
		if tt.name == "filepath-windows" && os.Getenv("GOOS") != "windows" {
			t.Skip("Cannot run windows test on a unix system")
		} else if tt.name == "filepath-unix" && os.Getenv("GOOS") != "linux" {
			t.Skip("Cannot run unix test on a windows system")
		}
		var want string
		got := SetError(tt.err)

		// if error has PII, string returned by SetError must not be the same as the error since it was sanitized
		// else it will be the same.
		if (tt.hasPII && got == tt.err.Error()) || (!tt.hasPII && got != tt.err.Error()) {
			if tt.hasPII {
				switch tt.name {
				case "username":
					want = strings.ReplaceAll(tt.err.Error(), user.Username, Sanitizer)
				case "filepath-unix":
					want = strings.ReplaceAll(tt.err.Error(), unixPath, Sanitizer)
				case "filepath-windows":
					want = strings.ReplaceAll(tt.err.Error(), windowsPath, Sanitizer)
				default:
				}
				t.Errorf("got: %q, want: %q", got, want)
			} else {
				t.Errorf("got: %s, want: %s", got, tt.err.Error())
			}
		}
	}
}

// createConfigDir creates a mock filesystem
func createConfigDir(t *testing.T) string {
	fs := filesystem.NewFakeFs()
	configDir, err := fs.TempDir(os.TempDir(), "telemetry")
	if err != nil {
		t.Error(err)
	}
	return configDir
}
