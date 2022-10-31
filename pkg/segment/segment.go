package segment

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Xuanwo/go-locale"

	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/pborman/uuid"
	"golang.org/x/term"
	"gopkg.in/segmentio/analytics-go.v3"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/preference"
)

// writekey will be the API key used to send data to the correct source on Segment. Default is the dev key
var writeKey = "4xGV1HV7K2FtUWaoAozSBD7SNCBCJ65U"

// Sanitizer replaces a PII data
const Sanitizer = "XXXX"

const TelemetryClient = "odo"

// DisableTelemetryEnv is name of environment variable, if set to true it disables odo telemetry completely
// hiding even the question
const (
	// DisableTelemetryEnv is name of environment variable, if set to true it disables odo telemetry completely.
	// Setting it to false has the same effect as not setting it at all == does NOT enable telemetry!
	// This has priority over TelemetryTrackingEnv
	//
	// Deprecated: Use TrackingConsentEnv instead.
	DisableTelemetryEnv = "ODO_DISABLE_TELEMETRY"
	// TrackingConsentEnv controls whether odo tracks telemetry or not.
	// Setting it to 'no' has the same effect as DisableTelemetryEnv=true (telemetry is disabled and no question asked)
	// Settings this to 'yes' skips the question about telemetry and enables user tracking.
	// Possible values are yes/no.
	TrackingConsentEnv    = "ODO_TRACKING_CONSENT"
	DebugTelemetryFileEnv = "ODO_DEBUG_TELEMETRY_FILE"
	TelemetryCaller       = "TELEMETRY_CALLER"
)

type TelemetryProperties struct {
	Duration      int64                  `json:"duration"`
	Error         string                 `json:"error"`
	ErrorType     string                 `json:"errortype"`
	Success       bool                   `json:"success"`
	Tty           bool                   `json:"tty"`
	Version       string                 `json:"version"`
	CmdProperties map[string]interface{} `json:"cmdProperties"`
}

type TelemetryData struct {
	Event      string              `json:"event"`
	Properties TelemetryProperties `json:"properties"`
}

type Client struct {
	// SegmentClient helps interact with the segment API
	SegmentClient analytics.Client
	// Preference points to the global odo config
	Preference preference.Client
	// TelemetryFilePath points to the file containing anonymousID used for tracking odo commands executed by the user
	TelemetryFilePath string
}

// NewClient returns a Client created with the default args
func NewClient(preference preference.Client) (*Client, error) {
	return newCustomClient(preference,
		GetTelemetryFilePath(),
		analytics.DefaultEndpoint,
	)
}

// newCustomClient returns a Client created with custom args
func newCustomClient(preference preference.Client, telemetryFilePath string, segmentEndpoint string) (*Client, error) {
	// get the locale information
	tag, err := locale.Detect()
	if err != nil {
		klog.V(4).Infof("couldn't fetch locale info: %s", err.Error())
	}
	// DefaultContext has IP set to 0.0.0.0 so that it does not track user's IP, which it does in case no IP is set
	client, err := analytics.NewWithConfig(writeKey, analytics.Config{
		Endpoint: segmentEndpoint,
		Verbose:  true,
		DefaultContext: &analytics.Context{
			IP:       net.IPv4(0, 0, 0, 0),
			Timezone: getTimeZoneRelativeToUTC(),
			OS: analytics.OSInfo{
				Name: runtime.GOOS,
			},
			Locale: tag.String(),
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

// getTimeZoneRelativeToUTC returns time zone relative to UTC (UTC +0530, UTC 0000, etc.)
func getTimeZoneRelativeToUTC() string {
	// t is a string array of RHC8222Z representation of current time
	// example - [13 Sep 21 16:37 +0530]
	var t = strings.Split(time.Now().Format(time.RFC822Z), " ")
	return fmt.Sprintf("UTC %s", t[len(t)-1])
}

// Close client connection and send the data
func (c *Client) Close() error {
	return c.SegmentClient.Close()
}

// Upload prepares the data to be sent to segment and send it once the client connection closes
func (c *Client) Upload(data TelemetryData) error {
	// if the user has not consented for telemetry, return
	if !IsTelemetryEnabled(c.Preference) {
		return nil
	}

	// obtain the user ID
	userId, uerr := getUserIdentity(c.TelemetryFilePath)
	if uerr != nil {
		return uerr
	}

	// add information to the data
	properties := analytics.NewProperties()
	for k, v := range data.Properties.CmdProperties {
		if k != scontext.TelemetryStatus {
			properties = properties.Set(k, v)
		}
	}

	properties = properties.Set("version", data.Properties.Version).
		Set("success", data.Properties.Success).
		Set("duration(ms)", data.Properties.Duration).
		Set("tty", data.Properties.Tty)
	// in case the command executed unsuccessfully, add information about the error in the data
	if data.Properties.Error != "" {
		properties = properties.Set("error", data.Properties.Error).Set("error-type", data.Properties.ErrorType)
	}

	// send the Identify message data that helps identify the user on segment
	err := c.SegmentClient.Enqueue(analytics.Identify{
		UserId: userId,
		Traits: addConfigTraits(),
	})
	if err != nil {
		klog.V(4).Infof("Cannot send Identify telemetry event: %q", err)
		// This doesn't have to be a fatal error, as we can still try to track normal event
		// There just might be some missing information about the user, but this will be only
		// in case that this was the first time we tried to send identify event for give userId.
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
	traits.Set("timezone", getTimeZoneRelativeToUTC())
	// get the locale information
	tag, err := locale.Detect()
	if err != nil {
		klog.V(4).Infof("couldn't fetch locale info: %s", err.Error())
	} else {
		traits.Set("locale", tag.String())
	}
	return traits
}

// GetTelemetryFilePath returns the default file path where the generated anonymous ID is stored
func GetTelemetryFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".redhat", "anonymousId")
}

// getUserIdentity returns the anonymous ID if it exists, else creates a new one and sends the data to Segment
func getUserIdentity(telemetryFilePath string) (string, error) {
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
	}
	return strings.TrimSpace(string(id)), nil
}

// SetError sanitizes any PII(Personally Identifiable Information) from the error
func SetError(err error) (errString string) {
	if err == nil {
		return ""
	}
	errString = err.Error()

	// Sanitize user information
	errString = sanitizeUserInfo(errString)

	// Sanitize file path
	errString = sanitizeFilePath(errString)

	// Sanitize exec commands: For errors when a command exec fails in cases like odo exec or odo test, we do not want to know the command that the user executed, so we simply return
	errString = sanitizeExec(errString)

	// Sanitize URL
	errString = sanitizeURL(errString)

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
func IsTelemetryEnabled(cfg preference.Client) bool {
	klog.V(4).Info("Checking telemetry enable status")
	// The env variable gets precedence in this decision.
	// In case a non-bool value was passed to the env var, we ignore it

	//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still need to support it
	disableTelemetry, _ := strconv.ParseBool(os.Getenv(DisableTelemetryEnv))
	if disableTelemetry {
		//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still need to support it
		klog.V(4).Infof("Sending telemetry disabled by %q env variable\n", DisableTelemetryEnv)
		return false
	}

	trackingConsentEnabled, present, err := IsTrackingConsentEnabled()
	if err != nil {
		klog.V(4).Infof("error in determining value of tracking consent env var: %v", err)
	} else if present {
		//Takes precedence over the ConsentTelemetry preference
		if !trackingConsentEnabled {
			klog.V(4).Info("Sending telemetry disabled by env variable\n")
			return false
		}
		klog.V(4).Info("Sending telemetry enabled by env variable\n")
		return true
	}

	isEnabled := cfg.GetConsentTelemetry()
	s := "Sending telemetry disabled by preference"
	if isEnabled {
		s = "Sending telemetry enabled by preference"
	}
	klog.V(4).Infof("%s\n", s)
	return isEnabled
}

// IsTrackingConsentEnabled returns whether tracking consent is enabled, based on the value of the TrackingConsentEnv environment variable.
// The second value returned indicates whether the variable is present in the environment.
func IsTrackingConsentEnabled() (enabled bool, present bool, err error) {
	trackingConsent, ok := os.LookupEnv(TrackingConsentEnv)
	if !ok {
		return false, false, nil
	}
	switch trackingConsent {
	case "yes":
		return true, true, nil
	case "no":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("invalid value for %s: %q", TrackingConsentEnv, trackingConsent)
	}
}

// sanitizeUserInfo sanitizes username from the error string
func sanitizeUserInfo(errString string) string {
	user1, err1 := user.Current()
	if err1 != nil {
		return err1.Error()
	}
	errString = strings.ReplaceAll(errString, user1.Username, Sanitizer)
	return errString
}

// sanitizeFilePath sanitizes file paths from error string
func sanitizeFilePath(errString string) string {
	for _, str := range strings.Split(errString, " ") {
		if strings.Count(str, string(os.PathSeparator)) > 1 {
			errString = strings.ReplaceAll(errString, str, Sanitizer)
		}
	}
	return errString
}

// sanitizeURL sanitizes URLs from the error string
func sanitizeURL(errString string) string {
	// the following regex parses hostnames and ip addresses
	// references - https://www.oreilly.com/library/view/regular-expressions-cookbook/9780596802837/ch07s16.html
	// https://www.oreilly.com/library/view/regular-expressions-cookbook/9781449327453/ch08s15.html
	urlPattern, err := regexp.Compile(`((https?|ftp|smtp)://)?((?:[0-9]{1,3}\.){3}[0-9]{1,3}(:([0-9]{1,5}))?|([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,})`)
	if err != nil {
		return errString
	}
	errString = urlPattern.ReplaceAllString(errString, Sanitizer)
	return errString
}

// sanitizeExec sanitizes commands from the error string that might have been executed by users while running commands like odo test or odo exec
func sanitizeExec(errString string) string {
	pattern, _ := regexp.Compile("exec command.*")
	errString = pattern.ReplaceAllString(errString, fmt.Sprintf("exec command %s", Sanitizer))
	return errString
}

func GetDebugTelemetryFile() string {
	return os.Getenv(DebugTelemetryFileEnv)
}
