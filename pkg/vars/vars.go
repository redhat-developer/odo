package vars

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// GetVariables returns a map of key/value from pairs defined in the file and in the list of strings in the format KEY=VALUE or KEY
// For a KEY entry, the value will be obtained from the environment, and if the value is not defined in the environment, the entry KEY will be ignored
// An empty filename will skip the extraction of pairs from file
func GetVariables(fs filesystem.Filesystem, filename string, override []string, lookupEnv func(string) (string, bool)) (map[string]string, error) {

	result := map[string]string{}
	var err error
	if len(filename) > 0 {
		result, err = parseKeyValueFile(fs, filename, lookupEnv)
		if err != nil {
			return nil, err
		}
	}
	overrideVars, err := parseKeyValueStrings(override, lookupEnv)
	if err != nil {
		return nil, err
	}

	for k, v := range overrideVars {
		result[k] = v
	}

	return result, nil
}

// parseKeyValueFile parses a file for "KEY=VALUE" lines and returns a map of keys/values
// If a key is defined without a value as "KEY", the value is searched into the environment with lookupEnv function
// Note that "KEY=" defines an empty value for KEY, but "KEY" indicates to search for value in environment
// If the KEY environment variable is not defined, this entry will be skipped
func parseKeyValueFile(fs filesystem.Filesystem, filename string, lookupEnv func(string) (string, bool)) (map[string]string, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := map[string]string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		scannedText := scanner.Text()

		key, value, err := parseKeyValueString(scannedText, lookupEnv)
		if err != nil {
			return nil, err
		}
		if len(key) == 0 {
			continue
		}
		result[key] = value
	}

	return result, nil
}

func parseKeyValueStrings(strs []string, lookupEnv func(string) (string, bool)) (map[string]string, error) {
	result := map[string]string{}

	for _, str := range strs {
		key, value, err := parseKeyValueString(str, lookupEnv)
		if err != nil {
			return nil, err
		}
		if len(key) == 0 {
			continue
		}
		result[key] = value
	}
	return result, nil
}

// parseKeyValueString parses a string to extract a key and its associated value
// if a line is empty or a comment, a nil error and an empty key are returned
// if a key does not define a value, the value will be obtained from the environment
// in this case, if the environment does not define the variable, the entry will be ignored
func parseKeyValueString(s string, lookupEnv func(string) (string, bool)) (string, string, error) {
	line := strings.TrimLeftFunc(s, unicode.IsSpace)
	if len(line) == 0 || strings.HasPrefix(line, "#") {
		return "", "", nil
	}
	parts := strings.SplitN(line, "=", 2)
	key := parts[0]

	// TODO validate key format

	if len(key) == 0 {
		return "", "", NewErrBadKey(fmt.Sprintf("no key defined in line %q", s))
	}

	var value string
	if len(parts) > 1 {
		value = parts[1]
	} else {
		var found bool
		value, found = lookupEnv(key)
		if !found {
			return "", "", nil
		}
	}

	return key, value, nil
}
