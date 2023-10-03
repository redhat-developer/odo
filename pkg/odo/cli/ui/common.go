package ui

import (
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"k8s.io/klog"
)

// Proceed displays a given message and asks the user if they want to proceed using the optionally specified Stdio instance (useful
// for testing purposes)
func Proceed(message string, stdio ...terminal.Stdio) (bool, error) {
	var response bool
	prompt := &survey.Confirm{
		Message: message,
	}

	if len(stdio) == 1 {
		prompt.WithStdio(stdio[0])
	}

	err := survey.AskOne(prompt, &response, survey.Required)
	if err != nil {
		klog.V(4).Infof("Encountered an error processing prompt: %v", err)
		if err == terminal.InterruptErr {
			return false, err
		}
		return false, nil
	}

	return response, nil
}
