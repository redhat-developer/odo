package segment

import (
	"context"
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
			Error         string `json:"error"`
			ErrorType     string `json:"error-type"`
			Success       bool   `json:"success"`
			Version       string `json:"version"`
			ComponentType string `json:"componentType"`
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

	testError := errors.New("error occurred")
	uploadData := fakeTelemetryData("odo preference view", testError, context.Background())
	// run a command, odo preference view
	if err = c.Upload(uploadData); err != nil {
		t.Error(err)
	}

	if err = c.Close(); err != nil {
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
			uploadData := fakeTelemetryData("odo create", tt.err, context.Background())
			// upload the data to Segment
			if err = c.Upload(uploadData); err != nil {
				t.Error(err)
			}
			// segment.Client.SegmentClient uploads the data to server when a condition is met or when the connection is closed.
			// This condition can be added by setting a BatchSize or Interval to the SegmentClient.
			// BatchSize or Interval conditions have not been set for segment.Client.SegmentClient, so we will need to
			// close the connection in order to upload the data to server.
			// In case a condition is added, we can close the client in the teardown.
			if err = c.Close(); err != nil {
				t.Error(err)
			}
			// Note: This will need to be changed if segment.Client.SegmentClient has BatchSize or Interval set to something.
			select {
			case x := <-body:
				s := segmentResponse{}
				if err1 := json.Unmarshal(x, &s); err1 != nil {
					t.Error(err1)
				}
				// Response returns 2 Batches in response -
				// 1) identify - user's system information in case the server did not already have this information,
				// and 2) track - information about the fired command
				// This condition checks if both the responses were received
				if s.Batch[0].Type != "identify" && s.Batch[1].Type != "track" {
					t.Errorf("Missing Identify or Track information.\nIdentify: %v\nTrack:%v", s.Batch[0].Type, s.Batch[1].Type)
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

func TestIsTelemetryEnabled(t *testing.T) {
	tests := []struct {
		errMesssage, envVar   string
		want, preferenceValue bool
	}{
		{
			want:            false,
			errMesssage:     "Telemetry must be disabled.",
			envVar:          "true",
			preferenceValue: false,
		},
		{
			want:            false,
			errMesssage:     "Telemetry must be disabled.",
			envVar:          "false",
			preferenceValue: false,
		},
		{
			want:            false,
			errMesssage:     "Telemetry must be disabled.",
			envVar:          "true",
			preferenceValue: true,
		},
		{
			want:            true,
			errMesssage:     "Telemetry must be enabled.",
			envVar:          "false",
			preferenceValue: true,
		},
		{
			want:            true,
			errMesssage:     "Telemetry must be enabled.",
			envVar:          "foobar",
			preferenceValue: true,
		},
		{
			want:            false,
			errMesssage:     "Telemetry must be disabled.",
			envVar:          "foobar",
			preferenceValue: false,
		},
	}
	for _, tt := range tests {
		os.Setenv(DisableTelemetryEnv, tt.envVar)
		cfg := &preference.PreferenceInfo{
			Preference: preference.Preference{
				OdoSettings: preference.OdoSettings{
					ConsentTelemetry: &tt.preferenceValue,
				},
			},
		}
		if IsTelemetryEnabled(cfg) != tt.want {
			t.Errorf(tt.errMesssage, "%s is set to %q. %s is set to %q.", DisableTelemetryEnv, tt.envVar, preference.ConsentTelemetrySetting, tt.preferenceValue)
		}
	}
}

func TestClientUploadWithContext(t *testing.T) {
	var uploadData TelemetryData
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
	ctx := NewContext(context.Background())

	for k, v := range map[string]string{"componentType": "nodejs"} {
		switch k {
		case "componentType":
			SetComponentType(ctx, v)
			uploadData = fakeTelemetryData("odo create", nil, ctx)
		}
		c, err := newCustomClient(cfg, createConfigDir(t), server.URL)
		if err != nil {
			t.Error(err)
		}
		// upload the data to Segment
		if err = c.Upload(uploadData); err != nil {
			t.Error(err)
		}
		if err = c.Close(); err != nil {
			t.Error(err)
		}
		select {
		case x := <-body:
			s := segmentResponse{}
			if err1 := json.Unmarshal(x, &s); err1 != nil {
				t.Error(err1)
			}
			if s.Batch[1].Type == "identify" {
				switch k {
				case "componentType":
					if s.Batch[1].Properties.ComponentType != v {
						t.Errorf("componentType did not match. Want: %q Got: %q", v, s.Batch[1].Properties.ComponentType)
					}

				}
			}
		default:
			t.Error("Server should receive some data")
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

// fakeTelemetryData returns fake data to test segment client Upload
func fakeTelemetryData(cmd string, err error, ctx context.Context) TelemetryData {
	return TelemetryData{
		Event: cmd,
		Properties: TelemetryProperties{
			Duration:      time.Second.Milliseconds(),
			Error:         SetError(err),
			ErrorType:     ErrorType(err),
			Success:       err == nil,
			Tty:           RunningInTerminal(),
			Version:       version.VERSION,
			CmdProperties: GetContextProperties(ctx),
		},
	}
}
