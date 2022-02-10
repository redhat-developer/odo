package ui

import (
	"os"

	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"k8s.io/klog"
)

// HandleError handles UI-related errors, in particular useful to gracefully handle ctrl-c interrupts gracefully
func HandleError(err error) {
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(1)
		} else {
			klog.V(4).Infof("Encountered an error processing prompt: %v", err)
		}
	}
}

// Proceed displays a given message and asks the user if they want to proceed using the optionally specified Stdio instance (useful
// for testing purposes)
func Proceed(message string, stdio ...terminal.Stdio) bool {
	var response bool
	prompt := &survey.Confirm{
		Message: message,
	}

	if len(stdio) == 1 {
		prompt.WithStdio(stdio[0])
	}

	err := survey.AskOne(prompt, &response, survey.Required)
	HandleError(err)

	return response
}
