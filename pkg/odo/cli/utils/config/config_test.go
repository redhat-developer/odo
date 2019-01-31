package config

import "testing"

func TestFormatSupportedParameters(t *testing.T) {
	expected := `
UpdateNotification - Controls if an update notification is shown or not (true or false)
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix
Timeout - Timeout (in seconds) for OpenShift server connection check`
	actual := formatSupportedParameters()
	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}
