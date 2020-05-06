package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Environment holds environment variables for new-app
type Environment map[string]string

// ParseEnvironmentAllowEmpty converts the provided strings in key=value form
// into environment entries. In case there's no equals sign in a string, it's
// considered as a key with empty value.
func ParseEnvironmentAllowEmpty(vals ...string) Environment {
	env := make(Environment)
	for _, s := range vals {
		if i := strings.Index(s, "="); i == -1 {
			env[s] = ""
		} else {
			env[s[:i]] = s[i+1:]
		}
	}
	return env
}

// ParseEnvironment takes a slice of strings in key=value format and transforms
// them into a map. List of duplicate keys is returned in the second return
// value.
func ParseEnvironment(vals ...string) (Environment, []string, []error) {
	errs := []error{}
	duplicates := []string{}
	env := make(Environment)
	for _, s := range vals {
		p := strings.SplitN(s, "=", 2)
		if err := validation.IsEnvVarName(p[0]); len(err) != 0 || len(p) != 2 {
			errs = append(errs, fmt.Errorf("invalid parameter assignment in %q, %v", s, err))
			continue
		}
		key, val := p[0], p[1]
		if _, exists := env[key]; exists {
			duplicates = append(duplicates, key)
			continue
		}
		env[key] = val
	}
	return env, duplicates, errs
}

// NewEnvironment returns a new set of environment variables based on all
// the provided environment variables
func NewEnvironment(envs ...map[string]string) Environment {
	if len(envs) == 1 {
		return envs[0]
	}
	out := make(Environment)
	out.Add(envs...)
	return out
}

// Add adds the environment variables to the current environment
func (e Environment) Add(envs ...map[string]string) {
	for _, env := range envs {
		for k, v := range env {
			e[k] = v
		}
	}
}

// AddIfNotPresent adds the environment variables to the current environment.
// In case of key conflict the old value is kept. Conflicting keys are returned
// as a slice.
func (e Environment) AddIfNotPresent(more Environment) []string {
	duplicates := []string{}
	for k, v := range more {
		if _, exists := e[k]; exists {
			duplicates = append(duplicates, k)
		} else {
			e[k] = v
		}
	}

	return duplicates
}

// List sorts and returns all the environment variables
func (e Environment) List() []corev1.EnvVar {
	env := []corev1.EnvVar{}
	for k, v := range e {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	sort.Sort(sortedEnvVar(env))
	return env
}

type sortedEnvVar []corev1.EnvVar

func (m sortedEnvVar) Len() int           { return len(m) }
func (m sortedEnvVar) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m sortedEnvVar) Less(i, j int) bool { return m[i].Name < m[j].Name }

// JoinEnvironment joins two different sets of environment variables
// into one, leaving out all the duplicates
func JoinEnvironment(a, b []corev1.EnvVar) (out []corev1.EnvVar) {
	out = a
	for i := range b {
		exists := false
		for j := range a {
			if a[j].Name == b[i].Name {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		out = append(out, b[i])
	}
	return out
}

// LoadEnvironmentFile accepts filename of a file containing key=value pairs
// and puts these pairs into a map. If filename is "-" the file contents are
// read from the stdin argument, provided it is not nil.
func LoadEnvironmentFile(filename string, stdin io.Reader) (Environment, error) {
	errorFilename := filename

	if filename == "-" && stdin != nil {
		//once https://github.com/joho/godotenv/pull/20 is merged we can get rid of using tempfile
		temp, err := ioutil.TempFile("", "origin-env-stdin")
		if err != nil {
			return nil, fmt.Errorf("Cannot create temporary file: %s", err)
		}

		filename = temp.Name()
		errorFilename = "stdin"
		defer os.Remove(filename)

		if _, err = io.Copy(temp, stdin); err != nil {
			return nil, fmt.Errorf("Cannot write to temporary file %q: %s", filename, err)
		}
		temp.Close()
	}

	// godotenv successfuly returns empty map when given path to a directory,
	// remove this once https://github.com/joho/godotenv/pull/22 is merged
	info, err := os.Stat(filename)
	if err == nil && info.IsDir() {
		return nil, fmt.Errorf("Cannot read variables from %q: is a directory", filename)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot stat %q: %s", filename, err)
	}

	size := info.Size()
	env, err := read(size, filename)
	if err != nil {
		return nil, fmt.Errorf("Cannot read variables from file %q: %s", errorFilename, err)
	}
	for k, v := range env {
		if err := validation.IsEnvVarName(k); len(err) != 0 {
			return nil, fmt.Errorf("invalid parameter assignment in %s=%s", k, v)
		}
	}
	return env, nil
}

// this is the start of a block of code that was copied from github.com/joho/godotenv;
// we diverted from that repo because the code did not account for files larger than 64K,
// and changes to their code we require either changes to the parameters on public methods/API
// or creation of methods like 'ReadWithSize' and 'ParseWithSize' to propoagate the size information;
// Convincing upstream of such klunky changes was untenable.
func read(size int64, filenames ...string) (envMap map[string]string, err error) {
	filenames = filenamesOrDefault(filenames)
	envMap = make(map[string]string)

	for _, filename := range filenames {
		individualEnvMap, individualErr := readFile(size, filename)

		if individualErr != nil {
			err = individualErr
			return // return early on a spazout
		}

		for key, value := range individualEnvMap {
			envMap[key] = value
		}
	}

	return
}
func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}

func readFile(size int64, filename string) (envMap map[string]string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return parse(size, file)
}

func parse(size int64, r io.Reader) (envMap map[string]string, err error) {
	envMap = make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(r)
	// changing the buf size is what we are changing from godotenv
	if size > bufio.MaxScanTokenSize {
		quotientOnly := size / bufio.MaxScanTokenSize
		// bump by one to move past any remainder
		quotientOnly++
		buf := make([]byte, quotientOnly*bufio.MaxScanTokenSize)
		scanner.Buffer(buf, int(quotientOnly*bufio.MaxScanTokenSize))
	}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine)

			if err != nil {
				return
			}
			envMap[key] = value
		}
	}
	return
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

func parseLine(line string) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(line, "=")
	firstColon := strings.Index(line, ":")
	splitString := strings.SplitN(line, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		//this is a yaml-style line
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.Trim(key, " ")

	// Parse the value
	value = parseValue(splitString[1])
	return
}

func parseValue(value string) string {

	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values or possible escapes
	if len(value) > 1 {
		first := string(value[0:1])
		last := string(value[len(value)-1:])
		if first == last && strings.ContainsAny(first, `"'`) {
			// pull the quotes off the edges
			value = value[1 : len(value)-1]
			// handle escapes
			escapeRegex := regexp.MustCompile(`\\.`)
			value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
				c := strings.TrimPrefix(match, `\`)
				switch c {
				case "n":
					return "\n"
				case "r":
					return "\r"
				default:
					return c
				}
			})
		}
	}

	return value
}

// this is the end of the block of code we copied from github.com/joho/godotenv
// that has fixes for creating Scanners on large files

// ParseAndCombineEnvironment parses key=value records from slice of strings
// (typically obtained from the command line) and from given files and combines
// them into single map. Key=value pairs from the envs slice have precedence
// over those read from file.
//
// The dupfn function is called for all duplicate keys that encountered. If the
// function returns an error this error is returned by
// ParseAndCombineEnvironment.
//
// If a file is "-" the file contents will be read from argument stdin (unless
// it's nil).
func ParseAndCombineEnvironment(envs []string, filenames []string, stdin io.Reader, dupfn func(string, string) error) (Environment, error) {
	vars, duplicates, errs := ParseEnvironment(envs...)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	for _, s := range duplicates {
		if err := dupfn(s, ""); err != nil {
			return nil, err
		}
	}

	for _, fname := range filenames {
		fileVars, err := LoadEnvironmentFile(fname, stdin)
		if err != nil {
			return nil, err
		}

		duplicates = vars.AddIfNotPresent(fileVars)
		for _, s := range duplicates {
			if err := dupfn(s, fname); err != nil {
				return nil, err
			}
		}
	}

	return vars, nil
}
