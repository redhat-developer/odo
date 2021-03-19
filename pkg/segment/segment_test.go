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
		AnonymousID string `json:"anonymousId"`
		MessageId   string `json:"messageId"`
		Traits      struct {
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
	// run a command, odo preference view
	if err = c.Upload("odo preference view", time.Second, errors.New("an error occurred")); err != nil {
		t.Error(err)
	}
	if err = c.Close(); err != nil {
		t.Error(err)
	}

	select {
	case <-body:
		t.Error("server should not receive data")
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
	}{
		{
			testName: "command ran successfully",
			err:      nil,
			success:  true,
			errType:  "",
		},
		{
			testName: "command failed",
			err:      errors.New("some error occurred"),
			success:  false,
			errType:  "*errors.errorString",
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			c, err := newCustomClient(cfg, createConfigDir(t), server.URL)
			if err != nil {
				t.Error(err)
			}
			//run a command, odo preference view
			if err := c.Upload("odo create", time.Second, tt.err); err != nil {
				t.Error(err)
			}
			if err := c.Close(); err != nil {
				t.Error(err)
			}

			select {
			case x := <-body:
				s := segmentResponse{}
				if err := json.Unmarshal(x, &s); err != nil {
					t.Error(err)
				}
				if s.Batch[0].Type != "identify" && s.Batch[1].Type != "track" {
					t.Error("Missing Identify or Track information")
				}
				if s.Batch[0].Traits.OS != runtime.GOOS {
					t.Error("OS does not match")
				}
				if !tt.success {
					if s.Batch[1].Properties.Error != tt.err.Error() {
						t.Error("Error does not match")
					}
				} else {
					if s.Batch[1].Properties.Error != "" {
						t.Error("Error does not match")
					}
				}
				if s.Batch[1].Properties.Success != tt.success {
					t.Error("Success does not match")
				}
				if s.Batch[1].Properties.ErrorType != tt.errType {
					t.Error("Error Type does not match")
				}
				if !strings.Contains(s.Batch[1].Properties.Version, version.VERSION) {
					t.Error("Odo version does not match")
				}

			default:
				t.Error("Server should receive data")
			}
		})
	}
}

func TestSetError(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error(err.Error())
	}

	tests := []struct {
		name   string
		err    error
		hasPII bool
	}{
		{
			name:   "error without PII information",
			err:    errors.New("this is an error string"),
			hasPII: false,
		},
		{
			name:   "error with PII information",
			err:    fmt.Errorf("cannot access the preference file '/home/%s/.odo/preference.yaml'", user.Username),
			hasPII: true,
		},
	}

	for _, tt := range tests {
		got := SetError(tt.err)

		// if error has PII, string returned by SetError must not be the same as the error since it was sanitized
		// else it will be the same.
		if (tt.hasPII && got == tt.err.Error()) || (!tt.hasPII && got != tt.err.Error()) {
			if tt.hasPII && strings.Contains(got, user.Username) {
				t.Error("PII was not sanitized properly.")
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
