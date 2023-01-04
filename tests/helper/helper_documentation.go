package helper

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	. "github.com/onsi/gomega"
)

const (
	timePatternInOdo = `(\[[0-9smh]+\])` // e.g. [4s], [1m], [3ms]
	staticTimeValue  = "[1s]"
	// Credit: https://github.com/acarl005/stripansi/blob/master/stripansi.go
	ansiPattern          = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	unicodeSpinnerFrames = "◓◐◑◒"
)

// ReplaceAllTimeInString replaces the time taken to download a Devfile or a starter project for an odo command with a custom value;
// this function is helpful because the time value is variable and replacing it with the value in mdx content helps in comparing.
func ReplaceAllTimeInString(docString string, timeString string) string {
	reg := regexp.MustCompile(timePatternInOdo)
	return reg.ReplaceAllString(docString, timeString)
}

// StripSpinner strips the cmd out string of spaces, spinner statements and spinner frames
func StripSpinner(docString string) (returnString string) {
	for _, line := range strings.Split(docString, "\n") {
		// trim any special character present in the line
		line = strings.TrimFunc(line, unicode.IsSpace)
		// This check is to avoid spinner statements in the cmd output
		// currently it does so for init and dev
		// e.g. " •  Syncing file changes ..."
		if (strings.HasPrefix(line, "•  Downloading") || strings.HasPrefix(line, "•  Syncing") || strings.HasPrefix(line, "•  Building")) && strings.HasSuffix(line, "...") {
			continue
		}

		// for some reason, splitting the docString by \n does not split the spinner frames,
		// so we perform a side operation to remove the extra spinner frames that are not present in the final output
		// e.g. "◐  Downloading devfile "java-maven" from registry "DefaultDevfileReg... \n ◓  Downloading devfile "java-maven" from registry "DefaultDevfileReg...\n\n [32m✓[0m  Downloading devfile "java-maven" from registry "DefaultDevfileRegistry" [2s]"
		if strings.ContainsAny(line, unicodeSpinnerFrames) {
			line = "✓" + strings.SplitAfter(line, "✓")[1]
		}

		returnString += line + "\n"
	}
	// replace all instances of time to [1s], this is also done for mdx out
	returnString = ReplaceAllTimeInString(returnString, staticTimeValue)
	return
}

// GetMDXContent reads the content of MDX files, strips it of extra spaces and returns the string
// it strips the extra space for an easy comparison
func GetMDXContent(filePath string) (mdxContent string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the docs directory
	mdxDir := filepath.Join(filepath.Dir(filename), "..", "..", "docs", "website", "docs")

	readFile, err := os.Open(filepath.Join(mdxDir, filePath))
	defer func(err error) {
		if err == nil {
			readFile.Close()
		}
	}(err)
	Expect(err).ToNot(HaveOccurred())

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimFunc(line, unicode.IsSpace)
		mdxContent += line + "\n"
	}

	// replace all instances of time to [1s], this is also done for cmd out
	mdxContent = ReplaceAllTimeInString(mdxContent, staticTimeValue)
	return
}

// StripAnsi strips the cmd out of ansi values used for fomatting(underline, colored line, etc.) the cmd out;
// this function should be called before StripSpinner for better results
// and is essential because mdx content does not support ansi
// The regex used by this function is copied from https://github.com/acarl005/stripansi/
func StripAnsi(docString string) (returnString string) {
	reg, err := regexp.Compile(ansiPattern)
	Expect(err).To(BeNil())
	returnString = reg.ReplaceAllString(docString, "")
	return
}

// StripInteractiveQuestion strips the extra output from interactive questions, leaving the final output
// e.g. "? Is this correct? (Y/n) No? Is this correct? No"
func StripInteractiveQuestion(docString string) (returnString string) {
	returnString = docString
	for _, question := range []string{"? Select language:", "? Select project type:", "? Select container for which you want to change configuration?", "? Is this correct?", "? Enter component name:", "? Which starter project do you want to use?"} {
		if strings.Count(returnString, question) > 1 {
			returnString = returnString[:strings.Index(returnString, question)] + returnString[strings.LastIndex(returnString, question):]
		}
	}
	return
}
