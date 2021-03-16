package segment

import (
	"errors"
	"fmt"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/version"
	"github.com/pborman/uuid"
	"golang.org/x/term"
	"gopkg.in/segmentio/analytics-go.v3"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var WriteKey = "CdhKrOlZ0YAtBz8OJXestPlp8CD2KkCc"

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

func (c *Client) Close() error {
	return c.segmentClient.Close()
}

func (c *Client) Upload(action string, duration time.Duration, err error) error {
	if !c.config.GetConsentTelemetry() {
		return nil
	}
	anonymousID, uerr := getUserIdentity(c.telemetryFilePath)
	if uerr != nil {
		return uerr
	}

	if err := c.segmentClient.Enqueue(analytics.Identify{
		AnonymousId: anonymousID,
		Traits:      addConfigTraits(c.config, traits()),
	}); err != nil {
		return err
	}

	properties := analytics.NewProperties()
	// TODO: add other properties when required
	//for k, v := range telemetry.GetContextProperties(ctx){
	//	properties = properties.Set(k, v)
	//}

	properties = properties.Set("version", "odo "+version.VERSION+" ("+version.GITCOMMIT+")").
		Set("success", err == nil).
		Set("duration(ms)", duration.Milliseconds()).
		Set("tty", RunningInTerminal())
	if err != nil {
		properties = properties.Set("error", SetError(err)).Set("error-type", errorType(err))
	}

	return c.segmentClient.Enqueue(analytics.Track{
		AnonymousId: anonymousID,
		Event:       action,
		Properties:  properties,
	})
}

func addConfigTraits(c *preference.PreferenceInfo, in analytics.Traits) analytics.Traits {
	// TODO: add traits later
	return in
}

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

func SetError(err error) string {
	// Sanitize user information
	user1, err1 := user.Current()
	if err1 != nil {
		return err1.Error()
	}
	return strings.ReplaceAll(err.Error(), user1.Username, "XXXX")
}

func errorType(err error) string {
	wrappedErr := errors.Unwrap(err)
	if wrappedErr != nil {
		return fmt.Sprintf("%T", wrappedErr)
	}
	return fmt.Sprintf("%T", err)
}

func RunningInTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
