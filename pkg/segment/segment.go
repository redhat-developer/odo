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
	"time"

	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/version"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"gopkg.in/segmentio/analytics-go.v3"
)

// writekey will be the API key used to send data to the correct source on Segment. Default is the dev key
const writeKey = "4xGV1HV7K2FtUWaoAozSBD7SNCBCJ65U"

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
func (c *Client) Upload(action string, duration time.Duration, err error) error {
	// if the user has not consented for telemetry, return
	if !c.Preference.GetConsentTelemetry() {
		return nil
	}

	// obtain the user ID
	userId, uerr := getUserIdentity(c.TelemetryFilePath)
	if uerr != nil {
		return uerr
	}

	// queue the data that helps identify the user on segment
	if err1 := c.SegmentClient.Enqueue(analytics.Identify{
		UserId: userId,
		Traits: addConfigTraits(),
	}); err1 != nil {
		return err1
	}

	// add information to the data
	properties := analytics.NewProperties()
	// TODO: add other properties when required
	properties = properties.Set("version", fmt.Sprintf("odo %v (%v)", version.VERSION, version.GITCOMMIT)).
		Set("success", err == nil).
		Set("duration(ms)", duration.Milliseconds()).
		Set("tty", RunningInTerminal())
	// in case the command executed unsuccessfully, add information about the error in the data
	if err != nil {
		properties = properties.Set("error", SetError(err)).Set("error-type", errorType(err))
	}

	// queue the data that has telemetry information
	return c.SegmentClient.Enqueue(analytics.Track{
		UserId:     userId,
		Event:      action,
		Properties: properties,
	})
}

// addConfigTraits adds information about the system
func addConfigTraits() analytics.Traits {
	traits := analytics.NewTraits().Set("os", runtime.GOOS)
	return traits
}

// getUserIdentity returns the anonymous ID if it exists, else creates a new one
func getUserIdentity(telemetryFilePath string) (string, error) {
	var id []byte

	// Get-or-Create the '$HOME/.redhat' directory
	if err := os.MkdirAll(filepath.Dir(telemetryFilePath), 0750); err != nil {
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
	}
	return strings.TrimSpace(string(id)), nil
}

// SetError sanitizes any PII(Personally Identifiable Information) from the error
func SetError(err error) string {
	// Sanitize user information
	user1, err1 := user.Current()
	if err1 != nil {
		return errors.Wrapf(err1, err1.Error()).Error()
	}
	return strings.ReplaceAll(err.Error(), user1.Username, "XXXX")
}

// errorType returns the type of error
func errorType(err error) string {
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
