package e2e

import (
	"os"
	"strings"
)

// createFileAtPath creates a file at the given path and writes the given content
// path is the path to the required file
// fileContent is the content to be written to the given file
func createFileAtPathWithContent(path string, fileContent string) error {
	// check if file exists
	var _, err = os.Stat(path)

	var file *os.File

	// create file if not exists
	if os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()

	} else {
		// Open file using READ & WRITE permission.
		file, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	// write to file
	_, err = file.WriteString(fileContent)
	if err != nil {
		return err
	}

	return nil
}

// determineRouteURL returns the http URL where the current component exposes it's service
// this URL can then be used in order to interact with the deployed service running in Openshift
// keeping with the spirit of the e2e tests, this expects, odo, sed and awk to be on the PATH
func determineRouteURL() string {
	output := runCmdShouldPass("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
	return strings.TrimSpace(output)
}
