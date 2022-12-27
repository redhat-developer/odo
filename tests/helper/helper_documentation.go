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
	ansi                 = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	unicodeSpinnerFrames = "◓◐◑◒"
)

func ReplaceTimeInString(docString string, timeString string) string {
	reg := regexp.MustCompile(timePatternInOdo)
	return reg.ReplaceAllString(docString, timeString)
}

func CleanStringOfSpinner(docString string) (returnString string) {
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

func GetMDXContent(filePath string) (mdxContent string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the docs directory
	mdxDir := filepath.Join(filepath.Dir(filename), "..", "..", "docs", "website", "docs")

	readFile, err := os.Open(filepath.Join(mdxDir, filePath))
	Expect(err).ToNot(HaveOccurred())
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimFunc(line, unicode.IsSpace)
		mdxContent += line + "\n"
	}
	return
}

func StripAnsi(docString string) (returnString string) {
	reg, err := regexp.Compile(ansi)
	Expect(err).To(BeNil())
	returnString = reg.ReplaceAllString(docString, "")
	return
}
