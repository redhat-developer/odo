package helper

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CompareDocOutput(cmdOut string, filePath string) ([]string, error) {
	var got = map[string]struct{}{}
	for _, line := range strings.Split(cmdOut, "\n") {
		if strings.Contains(line, "...") || line == "" {
			continue
		}
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
		return nil, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	var errString []string
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.Contains(line, "```") || strings.Contains(line, "$") || line == "" {
			continue
		}
		if _, ok := got[line]; !ok {
			// match partially, if cannot match exactly
			if !strings.Contains(cmdOut, line) {
				errString = append(errString, line)
			}
		}
	}
	return errString, nil
}
