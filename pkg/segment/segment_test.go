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

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/version"
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
			ClusterType   string `json:"clusterType"`
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

	ctrl := gomock.NewController(t)
	cfg := preference.NewMockClient(ctrl)
	cfg.EXPECT().GetConsentTelemetry().Return(false)

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
			ctrl := gomock.NewController(t)
			cfg := preference.NewMockClient(ctrl)
			cfg.EXPECT().GetConsentTelemetry().Return(true)

			c, err := newCustomClient(cfg, createConfigDir(t), server.URL)
			if err != nil {
				t.Error(err)
			}
			uploadData := fakeTelemetryData("odo init", tt.err, context.Background())
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

func TestIsTelemetryEnabled(t *testing.T) {
	type testStruct struct {
		name                 string
		env                  map[string]string
		consentTelemetryPref bool
		want                 func(odoDisableTelemetry, odoTrackingConsent string, consentTelemetry bool) bool
	}
	var tests []testStruct
	for _, odoDisableTelemetry := range []string{"", "true", "false", "foo"} {
		for _, odoTrackingConsent := range []string{"", "yes", "no", "bar"} {
			for _, consentTelemetry := range []bool{true, false} {
				odoDisableTelemetry := odoDisableTelemetry
				odoTrackingConsent := odoTrackingConsent
				consentTelemetry := consentTelemetry
				tests = append(tests, testStruct{
					want: func(odoDisableTelemetry, odoTrackingConsent string, consentTelemetry bool) bool {
						if odoDisableTelemetry == "true" || odoTrackingConsent == "no" {
							return false
						}
						if odoTrackingConsent == "yes" {
							return true
						}
						return consentTelemetry
					},
					name: fmt.Sprintf("ODO_DISABLE_TELEMETRY=%q,ODO_TRACKING_CONSENT=%q,ConsentTelemetry=%v",
						odoDisableTelemetry, odoTrackingConsent, consentTelemetry),
					env: map[string]string{
						//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still want to test it
						DisableTelemetryEnv: odoDisableTelemetry,
						TrackingConsentEnv:  odoTrackingConsent,
					},
					consentTelemetryPref: consentTelemetry,
				})
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			ctrl := gomock.NewController(t)
			cfg := preference.NewMockClient(ctrl)
			cfg.EXPECT().GetConsentTelemetry().Return(tt.consentTelemetryPref).AnyTimes()

			//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still want to test it
			if IsTelemetryEnabled(cfg) != tt.want(tt.env[DisableTelemetryEnv], tt.env[TrackingConsentEnv], tt.consentTelemetryPref) {
				t.Errorf(tt.name, "env is set to %v. %s is set to %q.", tt.env, preference.ConsentTelemetrySetting, tt.consentTelemetryPref)
			}
		})
	}
}

func TestClientUploadWithContext(t *testing.T) {
	var uploadData TelemetryData
	body, server := mockServer()
	defer server.Close()
	defer close(body)

	ctrl := gomock.NewController(t)
	cfg := preference.NewMockClient(ctrl)
	cfg.EXPECT().GetConsentTelemetry().Return(true).AnyTimes()
	ctx := scontext.NewContext(context.Background())

	for k, v := range map[string]string{scontext.ComponentType: "nodejs", scontext.ClusterType: ""} {
		switch k {
		case scontext.ComponentType:
			scontext.SetComponentType(ctx, v)
			uploadData = fakeTelemetryData("odo init", nil, ctx)
		case scontext.ClusterType:
			fakeClient, _ := kclient.FakeNew()
			scontext.SetClusterType(ctx, fakeClient)
			uploadData = fakeTelemetryData("odo set project", nil, ctx)
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
				case scontext.ComponentType:
					if s.Batch[1].Properties.ComponentType != v {
						t.Errorf("%v did not match. Want: %q Got: %q", scontext.ComponentType, v, s.Batch[1].Properties.ComponentType)
					}
				case scontext.ClusterType:
					if s.Batch[1].Properties.ClusterType != v {
						t.Errorf("%v did not match. Want: %q Got: %q", scontext.ClusterType, v, s.Batch[1].Properties.ClusterType)
					}
				}
			}
		default:
			t.Error("Server should receive some data")
		}
	}
}

func TestSetError(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error(err.Error())
	}

	tests := []struct {
		err    error
		hasPII bool
	}{
		{
			err:    errors.New("this is an error string"),
			hasPII: false,
		},
		{
			err:    fmt.Errorf("failed to execute devfile commands for component %s-comp. failed to Get https://my-cluster.project.local cannot run exec command [curl https://mycluster.domain.local -u foo -p password 123]", user.Username),
			hasPII: true,
		},
	}

	for _, tt := range tests {
		var want string
		got := SetError(tt.err)

		// if error has PII, string returned by SetError must not be the same as the error since it was sanitized
		// else it will be the same.
		if tt.hasPII {
			want = fmt.Sprintf("failed to execute devfile commands for component %s-comp. failed to Get %s cannot run exec command %s", Sanitizer, Sanitizer, Sanitizer)
		} else {
			want = tt.err.Error()
		}
		if got != want {
			t.Errorf("got: %q\nwant:%q", got, want)
		}

	}
}

func Test_sanitizeExec(t *testing.T) {
	err := fmt.Errorf("unable to execute the run command: unable to exec command [curl -K localhost:8080 -u user1 -p pwd123]")
	got := sanitizeExec(err.Error())
	want := fmt.Sprintf("unable to execute the run command: unable to exec command %s", Sanitizer)
	if got != want {
		t.Errorf("got: %q\nwant:%q", got, want)
	}
}

func Test_sanitizeURL(t *testing.T) {
	cases := []error{
		fmt.Errorf("resource project validation check failed.: Get https://my-cluster.project.local request cancelled"),
		fmt.Errorf("resource project validation check failed.: Get http://my-cluster.project.local request cancelled"),
		fmt.Errorf("resource project validation check failed.: Get http://192.168.0.1:6443 request cancelled"),
		fmt.Errorf("resource project validation check failed.: Get 10.18.25.1 request cancelled"),
		fmt.Errorf("resource project validation check failed.: Get www.sample.com request cancelled"),
	}

	for _, err := range cases {
		got := sanitizeURL(err.Error())
		want := fmt.Sprintf("resource project validation check failed.: Get %s request cancelled", Sanitizer)
		if got != want {
			t.Errorf("got: %q\nwant:%q", got, want)
		}
	}
}

func Test_sanitizeFilePath(t *testing.T) {
	unixPath := "/home/xyz/.odo/preference.yaml"
	windowsPath := "C:\\User\\XYZ\\preference.yaml"

	cases := []struct {
		name string
		err  error
	}{
		{
			name: "filepath-unix",
			err:  fmt.Errorf("cannot find the preference file at %s", unixPath),
		},
		{
			name: "filepath-windows",
			err:  fmt.Errorf("cannot find the preference file at %s", windowsPath),
		},
	}
	for _, tt := range cases {
		if tt.name == "filepath-windows" && os.Getenv("GOOS") != "windows" {
			t.Skip("Cannot run a windows test on a unix system")
		} else if tt.name == "filepath-unix" && os.Getenv("GOOS") != "linux" {
			t.Skip("Cannot run a unix test on a windows system")
		}

		got := sanitizeFilePath(tt.err.Error())
		want := fmt.Sprintf("cannot find the preference file at %s", Sanitizer)
		if got != want {
			t.Errorf("got: %q\nwant:%q", got, want)
		}
	}
}

func Test_sanitizeUserInfo(t *testing.T) {
	user, err1 := user.Current()
	if err1 != nil {
		t.Error(err1.Error())
	}

	err := fmt.Errorf("cannot create component name with %s", user.Username)
	got := sanitizeUserInfo(err.Error())
	want := fmt.Sprintf("cannot create component name with %s", Sanitizer)
	if got != want {
		t.Errorf("got: %q\nwant:%q", got, want)
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
			CmdProperties: scontext.GetContextProperties(ctx),
		},
	}
}
