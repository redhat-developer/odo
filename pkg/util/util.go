package util

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"k8s.io/klog"
)

var httpCacheDir = filepath.Join(os.TempDir(), "odohttpcache")

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

// ConvertLabelsToSelector converts the given labels to selector
// To pass operands such as !=, append a ! prefix to the value.
// For E.g. map[string]string{"app.kubernetes.io/managed-by": "!odo"}
// Using != operators also means that resource will be filtered even if it doesn't have the key.
// So a resource not labelled with key "app.kubernetes.io/managed-by" will also be returned.
// TODO(feloy) sync with devfile library?
func ConvertLabelsToSelector(labels map[string]string) string {
	var selector string
	isFirst := true
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := labels[k]
		if isFirst {
			isFirst = false
			if v == "" {
				selector = selector + fmt.Sprintf("%v", k)
			} else {
				if strings.HasPrefix(v, "!") {
					v = strings.Replace(v, "!", "", -1)
					selector = fmt.Sprintf("%v!=%v", k, v)
				} else {
					selector = fmt.Sprintf("%v=%v", k, v)
				}
			}
		} else {
			if v == "" {
				selector = selector + fmt.Sprintf(",%v", k)
			} else {
				if strings.HasPrefix(v, "!") {
					v = strings.Replace(v, "!", "", -1)
					selector = selector + fmt.Sprintf(",%v!=%v", k, v)
				} else {
					selector = selector + fmt.Sprintf(",%v=%v", k, v)
				}
			}
		}
	}
	return selector
}

// NamespaceKubernetesObject hyphenates applicationName and componentName
func NamespaceKubernetesObject(componentName string, applicationName string) (string, error) {

	// Error if it's blank
	if componentName == "" {
		return "", errors.New("namespacing: component name cannot be blank")
	}

	// Error if it's blank
	if applicationName == "" {
		return "", errors.New("namespacing: application name cannot be blank")
	}

	// Return the hyphenated namespaced name
	return fmt.Sprintf("%s-%s", strings.Replace(componentName, "/", "-", -1), applicationName), nil
}

// NamespaceKubernetesObjectWithTrim hyphenates applicationName and componentName
// if the resultant name is greater than 63 characters
// it trims each to 31 characters
// <31-characters>+"-"+<31-characters> = 63 characters
func NamespaceKubernetesObjectWithTrim(componentName, applicationName string) (string, error) {
	value, err := NamespaceKubernetesObject(componentName, applicationName)
	if err != nil {
		return "", err
	}

	// doesn't require trim
	if len(value) <= 63 {
		return value, nil
	}

	// trim to 31 characters
	componentName = componentName[:31]
	applicationName = applicationName[:31]
	value, err = NamespaceKubernetesObject(componentName, applicationName)
	if err != nil {
		return "", err
	}
	return value, nil
}

// TruncateString truncates passed string to given length
// Note: if -1 is passed, the original string is returned
// if appendIfTrunicated is given, then it will be appended to trunicated
// string
// TODO(feloy) sync with devfile library?
func TruncateString(str string, maxLen int, appendIfTrunicated ...string) string {
	if maxLen == -1 {
		return str
	}
	if len(str) > maxLen {
		truncatedString := str[:maxLen]
		for _, item := range appendIfTrunicated {
			truncatedString = fmt.Sprintf("%s%s", truncatedString, item)
		}
		return truncatedString
	}
	return str
}

// GetDNS1123Name Converts passed string into DNS-1123 string
// TODO(feloy) sync with devfile library?
func GetDNS1123Name(str string) string {
	nonAllowedCharsRegex := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	withReplacedChars := strings.Replace(
		nonAllowedCharsRegex.ReplaceAllString(str, "-"),
		"--", "-", -1)
	name := strings.ToLower(removeNonAlphaSuffix(removeNonAlphaPrefix(withReplacedChars)))
	// if the name is all numeric
	if len(str) != 0 && len(name) == 0 {
		name = strings.ToLower(removeNonAlphaSuffix(removeNonAlphaPrefix("x" + withReplacedChars)))
	}
	return name
}

func removeNonAlphaPrefix(input string) string {
	regex := regexp.MustCompile("^[^a-zA-Z]+(.*)$")
	return regex.ReplaceAllString(input, "$1")
}

func removeNonAlphaSuffix(input string) string {
	suffixRegex := regexp.MustCompile("^(.*?)[^a-zA-Z0-9]+$") // regex that strips all trailing non alpha-numeric chars
	matches := suffixRegex.FindStringSubmatch(input)
	matchesLength := len(matches)
	if matchesLength == 0 {
		// in this case the string does not contain a non-alphanumeric suffix
		return input
	}
	// in this case we return the smallest match which in the last element in the array
	return matches[matchesLength-1]
}

// CheckPathExists checks if a path exists or not
// TODO(feloy) use from devfile library?
func CheckPathExists(path string) bool {
	if _, err := filesystem.Get().Stat(path); !os.IsNotExist(err) {
		// path to file does exist
		return true
	}
	klog.V(4).Infof("path %s doesn't exist, skipping it", path)
	return false
}

// IsValidProjectDir checks that the folder to download the project from devfile is
// either empty or contains the devfile used.
// TODO(feloy) sync with devfile library?
func IsValidProjectDir(path string, devfilePath string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if len(files) >= 1 {
		for _, file := range files {
			fileName := file.Name()
			devfilePath = strings.TrimPrefix(devfilePath, "./")
			if !file.IsDir() && fileName == devfilePath {
				return nil
			}
		}
		return fmt.Errorf("Folder %s doesn't contain the devfile used.", path)
	}

	return nil
}

// GetAndExtractZip downloads a zip file from a URL with a http prefix or
// takes an absolute path prefixed with file:// and extracts it to a destination.
// pathToUnzip specifies the path within the zip folder to extract
// TODO(feloy) sync with devfile library?
func GetAndExtractZip(zipURL string, destination string, pathToUnzip string, starterToken string) error {
	if zipURL == "" {
		return fmt.Errorf("Empty zip url: %s", zipURL)
	}

	var pathToZip string
	if strings.HasPrefix(zipURL, "file://") {
		pathToZip = strings.TrimPrefix(zipURL, "file:/")
		if runtime.GOOS == "windows" {
			pathToZip = strings.Replace(pathToZip, "\\", "/", -1)
		}
	} else if strings.HasPrefix(zipURL, "http://") || strings.HasPrefix(zipURL, "https://") {
		// Generate temporary zip file location
		time := time.Now().Format(time.RFC3339)
		time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
		pathToZip = path.Join(os.TempDir(), "_"+time+".zip")

		params := dfutil.DownloadParams{
			Request: dfutil.HTTPRequestParams{
				URL:   zipURL,
				Token: starterToken,
			},
			Filepath: pathToZip,
		}
		err := dfutil.DownloadFile(params)
		if err != nil {
			return err
		}

		defer func() {
			if err := dfutil.DeletePath(pathToZip); err != nil {
				klog.Errorf("Could not delete temporary directory for zip file. Error: %s", err)
			}
		}()
	} else {
		return fmt.Errorf("Invalid Zip URL: %s . Should either be prefixed with file://, http:// or https://", zipURL)
	}

	filenames, err := Unzip(pathToZip, destination, pathToUnzip)
	if err != nil {
		return err
	}

	if len(filenames) == 0 {
		return errors.New("no files were unzipped, ensure that the project repo is not empty or that subDir has a valid path")
	}

	return nil
}

// Unzip will decompress a zip archive, moving specified files and folders
// within the zip file (parameter 1) to an output directory (parameter 2)
// Source: https://golangcode.com/unzip-files-in-go/
// pathToUnzip (parameter 3) is the path within the zip folder to extract
// TODO(feloy) sync with devfile library?
func Unzip(src, dest, pathToUnzip string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	// change path separator to correct character
	pathToUnzip = filepath.FromSlash(pathToUnzip)

	// removes first slash of pathToUnzip if present
	pathToUnzip = strings.TrimPrefix(pathToUnzip, string(os.PathSeparator))

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		index := strings.Index(f.Name, "/")
		filename := filepath.FromSlash(f.Name[index+1:])
		if filename == "" {
			continue
		}

		// if subDir has a pattern
		match, err := filepath.Match(pathToUnzip, filename)
		if err != nil {
			return filenames, err
		}

		// destination filepath before trim
		fpath := filepath.Join(dest, filename)

		// used for pattern matching
		fpathDir := filepath.Dir(fpath)

		// check for prefix or match
		if strings.HasPrefix(filename, pathToUnzip) {
			filename = strings.TrimPrefix(filename, pathToUnzip)
		} else if !strings.HasPrefix(filename, pathToUnzip) && !match && !sliceContainsString(fpathDir, filenames) {
			continue
		}
		// adds trailing slash to destination if needed as filepath.Join removes it
		if (len(filename) == 1 && os.IsPathSeparator(filename[0])) || filename == "" {
			fpath = dest + string(os.PathSeparator)
		} else {
			fpath = filepath.Join(dest, filename)
		}
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return filenames, err
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		// limit the number of bytes copied from a file
		// This is set to the limit of file size in Github
		// which is 100MB
		limited := io.LimitReader(rc, 100*1024*1024)

		_, err = io.Copy(outFile, limited)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// DownloadFileInMemory uses the url to download the file and return bytes
// TODO(feloy): sync with devfile library?
func DownloadFileInMemory(params dfutil.HTTPRequestParams) ([]byte, error) {
	data, err := dfutil.HTTPGetRequest(params, 0)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// DownloadFileInMemoryWithCache uses the url to download the file and return bytes
func DownloadFileInMemoryWithCache(params dfutil.HTTPRequestParams, cacheFor int) ([]byte, error) {
	data, err := dfutil.HTTPGetRequest(params, cacheFor)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// ValidateURL validates the URL
// TODO(feloy) sync with devfile library?
func ValidateURL(sourceURL string) error {
	// Valid URL needs to satisfy the following requirements:
	// 1. URL has scheme and host components
	// 2. Scheme, host of the URL shouldn't contain reserved character
	url, err := url.ParseRequestURI(sourceURL)
	if err != nil {
		return err
	}
	host := url.Host

	re := regexp.MustCompile(`[:\/\?#\[\]@]`)
	if host == "" || re.MatchString(host) {
		return errors.New("URL is invalid")
	}

	return nil
}

// GetDataFromURI gets the data from the given URI
// if the uri is a local path, we use the componentContext to complete the local path
func GetDataFromURI(uri, componentContext string, fs devfilefs.Filesystem) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if len(parsedURL.Host) != 0 && len(parsedURL.Scheme) != 0 {
		params := dfutil.HTTPRequestParams{
			URL: uri,
		}
		dataBytes, err := DownloadFileInMemoryWithCache(params, 1)
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	} else {
		dataBytes, err := fs.ReadFile(filepath.Join(componentContext, uri))
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	}
}

// sliceContainsString checks for existence of given string in given slice
func sliceContainsString(str string, slice []string) bool {
	for _, b := range slice {
		if b == str {
			return true
		}
	}
	return false
}

func addFileToIgnoreFile(gitIgnoreFile, filename string, fs filesystem.Filesystem) error {
	var data []byte
	file, err := fs.OpenFile(gitIgnoreFile, os.O_APPEND|os.O_RDWR, dfutil.ModeReadWriteFile)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore file: %w", err)
	}
	defer file.Close()

	if data, err = fs.ReadFile(gitIgnoreFile); err != nil {
		return fmt.Errorf("failed reading data from %v file: %w", gitIgnoreFile, err)
	}
	// check whether .odo/odo-file-index.json is already in the .gitignore file
	if !strings.Contains(string(data), filename) {
		if _, err := file.WriteString("\n" + filename); err != nil {
			return fmt.Errorf("failed to add %v to %v file: %w", filepath.Base(filename), gitIgnoreFile, err)
		}
	}
	return nil
}

// DisplayLog displays logs to user stdout with some color formatting
// numberOfLastLines limits the number of lines from the output when we are not following it
// TODO(feloy) sync with devfile library?
func DisplayLog(followLog bool, rd io.ReadCloser, writer io.Writer, compName string, numberOfLastLines int) (err error) {

	defer rd.Close()

	// Copy to stdout (in yellow)
	color.Set(color.FgYellow)
	defer color.Unset()

	// If we are going to followLog, we'll be copying it to stdout
	// else, we copy it to a buffer
	if followLog {

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			color.Unset()
			os.Exit(1)
		}()

		if _, err = io.Copy(writer, rd); err != nil {
			return fmt.Errorf("error followLoging logs for %s: %w", compName, err)
		}

	} else if numberOfLastLines == -1 {
		// Copy to buffer (we aren't going to be followLoging the logs..)
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rd)
		if err != nil {
			return fmt.Errorf("unable to copy followLog to buffer: %w", err)
		}

		// Copy to stdout
		if _, err = io.Copy(writer, buf); err != nil {
			return fmt.Errorf("error copying logs to stdout: %w", err)
		}
	} else {
		reader := bufio.NewReader(rd)
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return err
				} else {
					break
				}
			}

			lines = append(lines, line)
		}

		index := len(lines) - numberOfLastLines
		if index < 0 {
			index = 0
		}

		for i := index; i < len(lines)-1; i++ {
			_, err := fmt.Fprintf(writer, lines[i])
			if err != nil {
				return err
			}
		}
	}
	return

}

// copyFileWithFs copies a single file from src to dst
func copyFileWithFs(src, dst string, fs filesystem.Filesystem) error {
	var err error
	var srcinfo os.FileInfo

	srcfd, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if e := srcfd.Close(); e != nil {
			klog.V(4).Infof("err occurred while closing file: %v", e)
		}
	}()

	dstfd, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if e := dstfd.Close(); e != nil {
			klog.V(4).Infof("err occurred while closing file: %v", e)
		}
	}()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = fs.Stat(src); err != nil {
		return err
	}
	return fs.Chmod(dst, srcinfo.Mode())
}

// copyDirWithFS copies a whole directory recursively
func copyDirWithFS(src string, dst string, fs filesystem.Filesystem) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = fs.Stat(src); err != nil {
		return err
	}

	if err = fs.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = fs.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDirWithFS(srcfp, dstfp, fs); err != nil {
				return err
			}
		} else {
			if err = copyFileWithFs(srcfp, dstfp, fs); err != nil {
				return err
			}
		}
	}
	return nil
}

// StartSignalWatcher watches for signals and handles the situation before exiting the program
func StartSignalWatcher(watchSignals []os.Signal, handle func(receivedSignal os.Signal)) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, watchSignals...)
	defer signal.Stop(signals)

	receivedSignal := <-signals
	handle(receivedSignal)
	// exit here to stop spinners from rotating
	os.Exit(1)
}

// cleanDir cleans the original folder during events like interrupted copy etc
// it leaves the given files behind for later use
func cleanDir(originalPath string, leaveBehindFiles map[string]bool, fs filesystem.Filesystem) error {
	// Open the directory.
	outputDirRead, err := fs.Open(originalPath)
	if err != nil {
		return err
	}

	// Call Readdir to get all files.
	outputDirFiles, err := outputDirRead.Readdir(0)
	if err != nil {
		return err
	}

	// Loop over files.
	for _, file := range outputDirFiles {
		if value, ok := leaveBehindFiles[file.Name()]; ok && value {
			continue
		}
		err = fs.RemoveAll(filepath.Join(originalPath, file.Name()))
		if err != nil {
			return err
		}
	}
	return err
}

// GitSubDir handles subDir for git components using the default filesystem
func GitSubDir(srcPath, destinationPath, subDir string) error {
	return gitSubDir(srcPath, destinationPath, subDir, filesystem.DefaultFs{})
}

// gitSubDir handles subDir for git components
func gitSubDir(srcPath, destinationPath, subDir string, fs filesystem.Filesystem) error {
	go StartSignalWatcher([]os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt}, func(_ os.Signal) {
		err := cleanDir(destinationPath, map[string]bool{
			"devfile.yaml": true,
		}, fs)
		if err != nil {
			klog.V(4).Infof("error %v occurred while calling handleInterruptedSubDir", err)
		}
		err = fs.RemoveAll(srcPath)
		if err != nil {
			klog.V(4).Infof("error %v occurred during temp folder clean up", err)
		}
	})

	err := func() error {
		// Open the directory.
		outputDirRead, err := fs.Open(filepath.Join(srcPath, subDir))
		if err != nil {
			return err
		}
		defer func() {
			if err1 := outputDirRead.Close(); err1 != nil {
				klog.V(4).Infof("err occurred while closing temp dir: %v", err1)

			}
		}()
		// Call Readdir to get all files.
		outputDirFiles, err := outputDirRead.Readdir(0)
		if err != nil {
			return err
		}

		// Loop over files.
		for outputIndex := range outputDirFiles {
			outputFileHere := outputDirFiles[outputIndex]

			// Get name of file.
			fileName := outputFileHere.Name()

			oldPath := filepath.Join(srcPath, subDir, fileName)

			if outputFileHere.IsDir() {
				err = copyDirWithFS(oldPath, filepath.Join(destinationPath, fileName), fs)
			} else {
				err = copyFileWithFs(oldPath, filepath.Join(destinationPath, fileName), fs)
			}

			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}
	return fs.RemoveAll(srcPath)
}

// GetCommandStringFromEnvs creates a string from the given environment variables
func GetCommandStringFromEnvs(envVars []v1alpha2.EnvVar) string {
	var setEnvVariable string
	for i, envVar := range envVars {
		if i == 0 {
			setEnvVariable = "export"
		}
		setEnvVariable = setEnvVariable + fmt.Sprintf(" %v=\"%v\"", envVar.Name, envVar.Value)
	}
	return setEnvVariable
}

// GetGitOriginPath gets the remote fetch URL from the given git repo
// if the repo is not a git repo, the error is ignored
func GetGitOriginPath(path string) string {
	open, err := git.PlainOpen(path)
	if err != nil {
		return ""
	}

	remotes, err := open.Remotes()
	if err != nil {
		return ""
	}

	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			if len(remote.Config().URLs) > 0 {
				// https://github.com/go-git/go-git/blob/db4233e9e8b3b2e37259ed4e7952faaed16218b9/config/config.go#L549-L550
				// the first URL is the fetch URL
				return remote.Config().URLs[0]
			}
		}
	}
	return ""
}

// BoolPtr returns pointer to passed boolean
func GetBoolPtr(b bool) *bool {
	return &b
}

// SafeGetBool returns the value of the bool pointer, or false if the pointer is nil
func SafeGetBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// GetAdler32Value returns an adler32 hash of a string on 8 hexadecimal characters
func GetAdler32Value(s string) string {
	return fmt.Sprintf("%08x", adler32.Checksum([]byte(s)))
}

// IsPortFree checks if the port on localhost is free to use
func IsPortFree(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	_ = listener.Addr().(*net.TCPAddr).Port
	err = listener.Close()
	return err == nil
}

//WriteToJSONFile writes a struct to json file
func WriteToJSONFile(c interface{}, filename string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal data: %w", err)
	}

	if err = CreateIfNotExists(filename); err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		return fmt.Errorf("unable to write data to file %v: %w", c, err)
	}

	return nil
}
