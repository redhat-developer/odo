package helper

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	timePatternInOdo = `(\[[0-9smh]+\])` // e.g. [4s], [1m], [3ms]
)

// CompareDocOutput compares the output from test and the mdx file line by line;
// it ignores the time output if any is present in the line
// the function returns a list strings missing from the cmd out and from the file along with an error
func CompareDocOutput(cmdOut string, filePath string) (stringsMissingFromCmdOut, stringsMissingFromFile []string, err error) {
	// store lines of the cmdOut in this map
	var got = map[string]struct{}{}
	for _, line := range strings.Split(cmdOut, "\n") {
		// trim any space present at the beginning or end of a line
		line = strings.TrimSpace(line)

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
		line := strings.TrimSpace(fileScanner.Text())

		// ignore the line that starts with: 1) a code block "```", 2) the command "$", or 3) an empty line
		if strings.Contains(line, "```") || strings.Contains(line, "$") || line == "" {
			continue
		}

		line = removeTimeIfExists(line)

		// check if the line can be retrieved from the cmdOut map
		if _, ok := got[line]; !ok {
			// match partially, if cannot match exactly; this is helpful in case of backslash characters present in the cmdOut lines
			if !strings.Contains(cmdOut, line) {
				stringsMissingFromFile = append(stringsMissingFromFile, line)
			}
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
