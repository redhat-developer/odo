package segment

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/openshift/odo/pkg/preference"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"gopkg.in/segmentio/analytics-go.v3"
)

// writekey will be the API key used to send data to the correct source on Segment. Default is the dev key
var writeKey = "4xGV1HV7K2FtUWaoAozSBD7SNCBCJ65U"

// Sanitizer replaces a PII data
const Sanitizer = "XXXX"

// DisableTelemetryEnv is name of environment variable, if set to true it disables odo telemetry completely
// hiding even the question
const DisableTelemetryEnv = "ODO_DISABLE_TELEMETRY"

type TelmetryProperties struct {
	Duration  int64  `json:"duration"`
	Error     string `json:"error"`
	ErrorType string `json:"errortype"`
	Success   bool   `json:"success"`
	Tty       bool   `json:"tty"`
	Version   string `json:"version"`
}

type TelemetryData struct {
	Event      string             `json:"event"`
	Properties TelmetryProperties `json:"properties"`
}

type Client struct {
	// SegmentClient helps interact with the segment API
	SegmentClient analytics.Client
	// Preference points to the global odo config
	Preference *preference.PreferenceInfo
	// TelemetryFilePath points to the file containing anonymousID used for tracking odo commands executed by the user
	TelemetryFilePath string
}

// NewClient returns a Client created with the default args
func NewClient(preference *preference.PreferenceInfo) (*Client, error) {
	homeDir, _ := os.UserHomeDir()
	return newCustomClient(preference,
		filepath.Join(homeDir, ".redhat", "anonymousId"),
		analytics.DefaultEndpoint,
	)
}

// newCustomClient returns a Client created with custom args
func newCustomClient(preference *preference.PreferenceInfo, telemetryFilePath string, segmentEndpoint string) (*Client, error) {
	// DefaultContext has IP set to 0.0.0.0 so that it does not track user's IP, which it does in case no IP is set
	client, err := analytics.NewWithConfig(writeKey, analytics.Config{
		Endpoint: segmentEndpoint,
		Verbose:  true,
		DefaultContext: &analytics.Context{
			IP: net.IPv4(0, 0, 0, 0),
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		SegmentClient:     client,
		Preference:        preference,
		TelemetryFilePath: telemetryFilePath,
	}, nil
}

// Close client connection and send the data
func (c *Client) Close() error {
	return c.SegmentClient.Close()
}

// Upload prepares the data to be sent to segment and send it once the client connection closes
func (c *Client) Upload(data TelemetryData) error {
	err := data.Properties.Error
	// if the user has not consented for telemetry, return
	if !IsTelemetryEnabled(c.Preference) {
		return nil
	}

	// obtain the user ID
	userId, uerr := getUserIdentity(c, c.TelemetryFilePath)
	if uerr != nil {
		return uerr
	}

	// add information to the data
	properties := analytics.NewProperties()
	properties = properties.Set("version", data.Properties.Version).
		Set("success", data.Properties.Success).
		Set("duration(ms)", data.Properties.Duration).
		Set("tty", data.Properties.Tty)
	// in case the command executed unsuccessfully, add information about the error in the data
	if err != "" {
		properties = properties.Set("error", data.Properties.Error).Set("error-type", data.Properties.ErrorType)
	}

	// queue the data that has telemetry information
	return c.SegmentClient.Enqueue(analytics.Track{
		UserId:     userId,
		Event:      data.Event,
		Properties: properties,
	})
}

// addConfigTraits adds information about the system
func addConfigTraits() analytics.Traits {
	traits := analytics.NewTraits().Set("os", runtime.GOOS)
	return traits
}

// getUserIdentity returns the anonymous ID if it exists, else creates a new one and sends the data to Segment
func getUserIdentity(client *Client, telemetryFilePath string) (string, error) {
	var id []byte

	// Get-or-Create the '$HOME/.redhat' directory
	if err := os.MkdirAll(filepath.Dir(telemetryFilePath), os.ModePerm); err != nil {
		return "", err
	}

	// Get-or-Create the anonymousID file that contains a UUID
	if _, err := os.Stat(telemetryFilePath); !os.IsNotExist(err) {
		id, err = ioutil.ReadFile(telemetryFilePath)
		if err != nil {
			return "", err
		}
	}

	// check if the id is a valid uuid, if not, nil is returned
	if uuid.Parse(strings.TrimSpace(string(id))) == nil {
		id = []byte(uuid.NewRandom().String())
		if err := ioutil.WriteFile(telemetryFilePath, id, 0600); err != nil {
			return "", err
		}
		// Since a new ID was created, send the Identify message data that helps identify the user on segment
		if err1 := client.SegmentClient.Enqueue(analytics.Identify{
			UserId: strings.TrimSpace(string(id)),
			Traits: addConfigTraits(),
		}); err1 != nil {
			// TODO: maybe change this to klog Info instead of returning
			return "", err1
		}

	}
	return strings.TrimSpace(string(id)), nil
}

// SetError sanitizes any PII(Personally Identifiable Information) from the error
func SetError(err error) (errString string) {
	if err == nil {
		return ""
	}
	// sanitize user information
	user1, err1 := user.Current()
	if err1 != nil {
		return errors.Wrapf(err1, err1.Error()).Error()
	}
	errString = strings.ReplaceAll(err.Error(), user1.Username, Sanitizer)

	// sanitize file path
	for _, str := range strings.Split(errString, " ") {
		if strings.Count(str, string(os.PathSeparator)) > 1 {
			errString = strings.ReplaceAll(errString, str, Sanitizer)
		}
	}
	return errString
}

// ErrorType returns the type of error
func ErrorType(err error) string {
	if err == nil {
		return ""
	}
	wrappedErr := errors.Unwrap(err)
	if wrappedErr != nil {
		return fmt.Sprintf("%T", wrappedErr)
	}
	return fmt.Sprintf("%T", err)
}

// RunningInTerminal checks if odo was run from a terminal
func RunningInTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// IsTelemetryEnabled returns true if user has consented to telemetry
func IsTelemetryEnabled(cfg *preference.PreferenceInfo) bool {
	// The env variable gets precedence over the decision
	if os.Getenv(DisableTelemetryEnv) == "true" {
		return false
	} else if cfg.GetConsentTelemetry() {
		return true
	}
	return false
}
