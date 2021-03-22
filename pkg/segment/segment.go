package segment

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/version"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"gopkg.in/segmentio/analytics-go.v3"
)

// Writekey will be the API key used to send data to the correct source on Segment
var WriteKey = "R1Z79HadJIrphLoeONZy5uqOjusljSwN"

type Client struct {
	segmentClient     analytics.Client
	config            *preference.PreferenceInfo
	telemetryFilePath string
}

func NewClient(config *preference.PreferenceInfo) (*Client, error) {
	homeDir, _ := os.UserHomeDir()
	return newCustomClient(config,
		filepath.Join(homeDir, ".redhat", "anonymousId"),
		analytics.DefaultEndpoint,
	)
}

func newCustomClient(config *preference.PreferenceInfo, telemetryFilePath string, segmentEndpoint string) (*Client, error) {
	client, err := analytics.NewWithConfig(WriteKey, analytics.Config{
		Endpoint: segmentEndpoint,
		DefaultContext: &analytics.Context{
			IP: net.IPv4(0, 0, 0, 0),
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		segmentClient:     client,
		config:            config,
		telemetryFilePath: telemetryFilePath,
	}, nil
}

// Close client connection and send the data
func (c *Client) Close() error {
	return c.segmentClient.Close()
}

// Upload prepares the data to be sent to segment and send it once the client connection closes
func (c *Client) Upload(action string, duration time.Duration, err error) error {
	// if the user has not consented for telemetry, return
	if !c.config.GetConsentTelemetry() {
		return nil
	}

	// obtain the anonymous ID
	anonymousID, uerr := getUserIdentity(c.telemetryFilePath)
	if uerr != nil {
		return uerr
	}

	// queue the data that helps identify the user on segment
	if err := c.segmentClient.Enqueue(analytics.Identify{
		AnonymousId: anonymousID,
		Traits:      addConfigTraits(c.config, traits()),
	}); err != nil {
		return err
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
	return c.segmentClient.Enqueue(analytics.Track{
		AnonymousId: anonymousID,
		Event:       action,
		Properties:  properties,
	})
}

// addConfigTraits add more information to be sent to segment
// Note: This currently acts as a placeholder
func addConfigTraits(c *preference.PreferenceInfo, in analytics.Traits) analytics.Traits {
	// TODO: add more traits later
	return in
}

// getUserIdentity returns the anonymous ID if it exists, else creates a new one
func getUserIdentity(telemetryFilePath string) (string, error) {
	var id []byte

	if err := os.MkdirAll(filepath.Dir(telemetryFilePath), 0750); err != nil {
		return "", err
	}

	if _, err := os.Stat(telemetryFilePath); !os.IsNotExist(err) {
		id, err = ioutil.ReadFile(telemetryFilePath)
		if err != nil {
			return "", err
		}
	}

	// check if the id a valid uuid, if it is not, nil is returned
	if uuid.Parse(strings.TrimSpace(string(id))) == nil {
		id = []byte(uuid.NewRandom().String())
		if err := ioutil.WriteFile(telemetryFilePath, id, 0600); err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(id)), nil
}

// SetError sanitizes any pii information from the error
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
