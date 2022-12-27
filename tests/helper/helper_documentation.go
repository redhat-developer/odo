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

// CompareDocOutput compares the output from test and the mdx file line by line;
// it ignores the time output if any is present in the line
// the function returns a list strings missing from the cmd out and from the file along with an error
// TODO: make cmdOut string same as mdx and compare them both
// TODO: can pass static values of time(can be something else as well) and pattern as defined in mdx while comparing
func CompareDocOutput(cmdOut string, filePath string) (stringsMissingFromCmdOut, stringsMissingFromFile []string, err error) {
	// store lines of the cmdOut in this map
	var got = map[string]struct{}{}
	for _, line := range strings.Split(cmdOut, "\n") {
		// trim any space present at the beginning or end of a line
		line = strings.TrimFunc(line, unicode.IsSpace)
		if strings.Contains(line, "...") || line == "" {
			continue
		}

		line = removeTimeIfExists(line)
		got[line] = struct{}{}
	}

	mdxDir := func() string {
		// filename of this file
		_, filename, _, _ := runtime.Caller(0)
		// path to the docs directory
		return filepath.Join(filepath.Dir(filename), "..", "..", "docs", "website", "docs")
	}()

	readFile, err := os.Open(filepath.Join(mdxDir, filePath))
	if err != nil {
		return nil, nil, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		// trim any space present at the beginning or end of a line
		line := strings.TrimFunc(fileScanner.Text(), unicode.IsSpace)

		// ignore the line that starts with: 1) a code block "```", 2) the command "$", or 3) an empty line
		if strings.Contains(line, "```") || strings.Contains(line, "$") || line == "" || strings.Contains(line, "...") {
			continue
		}

		line = removeTimeIfExists(line)

		// check if the line can be retrieved from the cmdOut map
		if _, ok := got[line]; !ok {
			stringsMissingFromFile = append(stringsMissingFromFile, line)
		} else {
			delete(got, line)
		}
	}

	for line := range got {
		stringsMissingFromCmdOut = append(stringsMissingFromCmdOut, line)
	}

	return stringsMissingFromCmdOut, stringsMissingFromFile, nil
}

// removeTimeIfExists removes time string from a line
// e.g. of a time string: [4s], [1m], [3ms]
func removeTimeIfExists(line string) string {
	// check if a line has time data by checking for closing bracket of [4s]
	if hasTimeDataInLine := strings.HasSuffix(line, "]"); !hasTimeDataInLine {
		return line
	}
	reg := regexp.MustCompile(timePatternInOdo)
	return reg.ReplaceAllString(line, "")
}

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
