package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintNoArgsWarningMessage(command string) {
	fmt.Printf(`
No arg received. Did you forget to add the project path to analyze?

Expected:
  alizer %s /your/local/project/path [flags]

To find out more, run 'alizer %s --help'
`, command, command)
}

func PrintWrongLoggingLevelMessage(command string) {
	fmt.Printf(`Argument log has wrong value. Did you choose one of debug, warning, info?

Expected:
  alizer %s /your/local/project/path --log [debug, warning, info]

To find out more, run 'alizer %s --help'
`, command, command)
}

func PrintPrettifyOutput(value interface{}, err error) {
	if err != nil {
		RedirectErrorToStdErrAndExit(err)
	}
	b, err1 := json.MarshalIndent(value, "", "\t")
	if err1 != nil {
		RedirectErrorToStdErrAndExit(err1)
	}
	fmt.Println(string(b))
}

func RedirectErrorToStdErrAndExit(err error) {
	RedirectErrorStringToStdErrAndExit(err.Error())
}

func RedirectErrorStringToStdErrAndExit(err string) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
