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
	// Credit: https://github.com/acarl005/stripansi/blob/master/stripansi.go
	ansiPattern          = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	unicodeSpinnerFrames = "◓◐◑◒"
)

// ReplaceTimeInString replaces the time taken to download a Devfile or a starter project for an odo command with a custom value;
// this function is helpful because the time value is variable and replacing it with the value in mdx content helps in comparing.
func ReplaceTimeInString(docString string, timeString string) string {
	reg := regexp.MustCompile(timePatternInOdo)
	return reg.ReplaceAllString(docString, timeString)
}

// StripSpinner strips the cmd out string of spaces, spinner statements and spinner frames
func StripSpinner(docString string) (returnString string) {
	for _, line := range strings.Split(docString, "\n") {
		// trim any special character present in the line
		line = strings.TrimFunc(line, unicode.IsSpace)
		// This check is to avoid spinner statements in the cmd output
		if strings.ContainsAny(line, unicodeSpinnerFrames) || strings.HasSuffix(line, "...") {
			continue
		}
		returnString += line + "\n"
	}
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
	defer readFile.Close()
	Expect(err).ToNot(HaveOccurred())

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimFunc(line, unicode.IsSpace)
		mdxContent += line + "\n"
	}
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
