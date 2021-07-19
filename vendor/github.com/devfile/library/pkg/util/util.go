package util

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
)

const (
	HTTPRequestTimeout    = 30 * time.Second // HTTPRequestTimeout configures timeout of all HTTP requests
	ResponseHeaderTimeout = 30 * time.Second // ResponseHeaderTimeout is the timeout to retrieve the server's response headers
	ModeReadWriteFile     = 0600             // default Permission for a file
	CredentialPrefix      = "odo-"           // CredentialPrefix is the prefix of the credential that uses to access secure registry
)

// httpCacheDir determines directory where odo will cache HTTP respones
var httpCacheDir = filepath.Join(os.TempDir(), "odohttpcache")

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

// 63 is the max length of a DeploymentConfig in Openshift and we also have to take into account
// that each component also gets a volume that uses the component name suffixed with -s2idata
const maxAllowedNamespacedStringLength = 63 - len("-s2idata") - 1

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

const defaultGithubRef = "master"

// ResourceRequirementInfo holds resource quantity before transformation into its appropriate form in container spec
type ResourceRequirementInfo struct {
	ResourceType corev1.ResourceName
	MinQty       resource.Quantity
	MaxQty       resource.Quantity
}

// HTTPRequestParams holds parameters of forming http request
type HTTPRequestParams struct {
	URL   string
	Token string
}

// DownloadParams holds parameters of forming file download request
type DownloadParams struct {
	Request  HTTPRequestParams
	Filepath string
}

// ConvertLabelsToSelector converts the given labels to selector
func ConvertLabelsToSelector(labels map[string]string) string {
	var selector string
	isFirst := true
	for k, v := range labels {
		if isFirst {
			isFirst = false
			if v == "" {
				selector = selector + fmt.Sprintf("%v", k)
			} else {
				selector = fmt.Sprintf("%v=%v", k, v)
			}
		} else {
			if v == "" {
				selector = selector + fmt.Sprintf(",%v", k)
			} else {
				selector = selector + fmt.Sprintf(",%v=%v", k, v)
			}
		}
	}
	return selector
}

// GenerateRandomString generates a random string of lower case characters of
// the given size
func GenerateRandomString(n int) string {
	b := make([]rune, n)

	for i := range b {
		// this error is ignored because it fails only when the 2nd arg of Int() is less then 0
		// which wont happen
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		b[i] = letterRunes[n.Int64()]
	}
	return string(b)
}

// In checks if the value is in the array
func In(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

// NamespaceOpenShiftObject hyphenates applicationName and componentName
func NamespaceOpenShiftObject(componentName string, applicationName string) (string, error) {

	// Error if it's blank
	if componentName == "" {
		return "", errors.New("namespacing: component name cannot be blank")
	}

	// Error if it's blank
	if applicationName == "" {
		return "", errors.New("namespacing: application name cannot be blank")
	}

	// Return the hyphenated namespaced name
	originalName := fmt.Sprintf("%s-%s", strings.Replace(componentName, "/", "-", -1), applicationName)
	truncatedName := TruncateString(originalName, maxAllowedNamespacedStringLength)
	if originalName != truncatedName {
		klog.V(4).Infof("The combination of application %s and component %s was too long so the final name was truncated to %s",
			applicationName, componentName, truncatedName)
	}
	return truncatedName, nil
}

// ExtractComponentType returns only component type part from passed component type(default unqualified, fully qualified, versioned, etc...and their combinations) for use as component name
// Possible types of parameters:
// 1. "myproject/python:3.5" -- Return python
// 2. "python:3.5" -- Return python
// 3. nodejs -- Return nodejs
func ExtractComponentType(namespacedVersionedComponentType string) string {
	s := strings.Split(namespacedVersionedComponentType, "/")
	versionedString := s[0]
	if len(s) == 2 {
		versionedString = s[1]
	}
	s = strings.Split(versionedString, ":")
	return s[0]
}

// ParseComponentImageName returns
// 1. image name
// 2. component type i.e, builder image name
// 3. component name default value is component type else the user requested component name
// 4. component version which is by default latest else version passed with builder image name
func ParseComponentImageName(imageName string) (string, string, string, string) {
	// We don't have to check it anymore, Args check made sure that args has at least one item
	// and no more than two

	// "Default" values
	componentImageName := imageName
	componentType := imageName
	componentName := ExtractComponentType(componentType)
	componentVersion := "latest"

	// Check if componentType includes ":", if so, then we need to spit it into using versions
	if strings.ContainsAny(componentImageName, ":") {
		versionSplit := strings.Split(imageName, ":")
		componentType = versionSplit[0]
		componentName = ExtractComponentType(componentType)
		componentVersion = versionSplit[1]
	}
	return componentImageName, componentType, componentName, componentVersion
}

// WIN represent the windows OS
const WIN = "windows"

// ReadFilePath Reads file path form URL file:///C:/path/to/file to C:\path\to\file
func ReadFilePath(u *url.URL, os string) string {
	location := u.Path
	if os == WIN {
		location = strings.Replace(u.Path, "/", "\\", -1)
		location = location[1:]
	}
	return location
}

// GenFileURL Converts file path on windows to /C:/path/to/file to work in URL
func GenFileURL(location string, os ...string) string {
	// param os is made variadic only for the purpose of UTs but need not be passed mandatorily
	currOS := runtime.GOOS
	if len(os) > 0 {
		currOS = os[0]
	}
	urlPath := location
	if currOS == WIN {
		urlPath = "/" + strings.Replace(location, "\\", "/", -1)
	}
	return "file://" + urlPath
}

// ConvertKeyValueStringToMap converts String Slice of Parameters to a Map[String]string
// Each value of the slice is expected to be in the key=value format
// Values that do not conform to this "spec", will be ignored
func ConvertKeyValueStringToMap(params []string) map[string]string {
	result := make(map[string]string, len(params))
	for _, param := range params {
		str := strings.Split(param, "=")
		if len(str) != 2 {
			klog.Fatalf("Parameter %s is not in the expected key=value format", param)
		} else {
			result[str[0]] = str[1]
		}
	}
	return result
}

// TruncateString truncates passed string to given length
// Note: if -1 is passed, the original string is returned
func TruncateString(str string, maxLen int) string {
	if maxLen == -1 {
		return str
	}
	if len(str) > maxLen {
		return str[:maxLen]
	}
	return str
}

// GetAbsPath returns absolute path from passed file path resolving even ~ to user home dir and any other such symbols that are only
// shell expanded can also be handled here
func GetAbsPath(path string) (string, error) {
	// Only shell resolves `~` to home so handle it specially
	var dir string
	if strings.HasPrefix(path, "~") {
		if len(customHomeDir) > 0 {
			dir = customHomeDir
		} else {
			usr, err := user.Current()
			if err != nil {
				return path, errors.Wrapf(err, "unable to resolve %s to absolute path", path)
			}
			dir = usr.HomeDir
		}

		if len(path) > 1 {
			path = filepath.Join(dir, path[1:])
		} else {
			path = dir
		}
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return path, errors.Wrapf(err, "unable to resolve %s to absolute path", path)
	}
	return path, nil
}

// GetRandomName returns a randomly generated name which can be used for naming odo and/or openshift entities
// prefix: Desired prefix part of the name
// prefixMaxLen: Desired maximum length of prefix part of random name; if -1 is passed, no limit on length will be enforced
// existList: List to verify that the returned name does not already exist
// retries: number of retries to try generating a unique name
// Returns:
//		1. randomname: is prefix-suffix, where:
//				prefix: string passed as prefix or fetched current directory of length same as the passed prefixMaxLen
//				suffix: 4 char random string
//      2. error: if requested number of retries also failed to generate unique name
func GetRandomName(prefix string, prefixMaxLen int, existList []string, retries int) (string, error) {
	prefix = TruncateString(GetDNS1123Name(strings.ToLower(prefix)), prefixMaxLen)
	name := fmt.Sprintf("%s-%s", prefix, GenerateRandomString(4))

	//Create a map of existing names for efficient iteration to find if the newly generated name is same as any of the already existing ones
	existingNames := make(map[string]bool)
	for _, existingName := range existList {
		existingNames[existingName] = true
	}

	// check if generated name is already used in the existList
	if _, ok := existingNames[name]; ok {
		prevName := name
		trial := 0
		// keep generating names until generated name is not unique. So, loop terminates when name is unique and hence for condition is false
		for ok {
			trial = trial + 1
			prevName = name
			// Attempt unique name generation from prefix-suffix by concatenating prefix-suffix withrandom string of length 4
			prevName = fmt.Sprintf("%s-%s", prevName, GenerateRandomString(4))
			_, ok = existingNames[prevName]
			if trial >= retries {
				// Avoid infinite loops and fail after passed number of retries
				return "", fmt.Errorf("failed to generate a unique name even after %d retrials", retries)
			}
		}
		// If found to be unique, set name as generated name
		name = prevName
	}
	// return name
	return name, nil
}

// GetDNS1123Name Converts passed string into DNS-1123 string
func GetDNS1123Name(str string) string {
	nonAllowedCharsRegex := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	withReplacedChars := strings.Replace(
		nonAllowedCharsRegex.ReplaceAllString(str, "-"),
		"--", "-", -1)
	return removeNonAlphaSuffix(removeNonAlphaPrefix(withReplacedChars))
}

func removeNonAlphaPrefix(input string) string {
	regex := regexp.MustCompile("^[^a-zA-Z0-9]+(.*)$")
	return regex.ReplaceAllString(input, "$1")
}

func removeNonAlphaSuffix(input string) string {
	suffixRegex := regexp.MustCompile("^(.*?)[^a-zA-Z0-9]+$") //regex that strips all trailing non alpha-numeric chars
	matches := suffixRegex.FindStringSubmatch(input)
	matchesLength := len(matches)
	if matchesLength == 0 {
		// in this case the string does not contain a non-alphanumeric suffix
		return input
	} else {
		// in this case we return the smallest match which in the last element in the array
		return matches[matchesLength-1]
	}
}

// SliceDifference returns the values of s2 that do not exist in s1
func SliceDifference(s1 []string, s2 []string) []string {
	mb := map[string]bool{}
	for _, x := range s1 {
		mb[x] = true
	}
	difference := []string{}
	for _, x := range s2 {
		if _, ok := mb[x]; !ok {
			difference = append(difference, x)
		}
	}
	return difference
}

// OpenBrowser opens the URL within the users default browser
func OpenBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}

	return nil
}

// FetchResourceQuantity takes passed min, max and requested resource quantities and returns min and max resource requests
func FetchResourceQuantity(resourceType corev1.ResourceName, min string, max string, request string) (*ResourceRequirementInfo, error) {
	if min == "" && max == "" && request == "" {
		return nil, nil
	}
	// If minimum and maximum both are passed they carry highest priority
	// Otherwise, use the request as min and max
	var minResource resource.Quantity
	var maxResource resource.Quantity
	if min != "" {
		resourceVal, err := resource.ParseQuantity(min)
		if err != nil {
			return nil, err
		}
		minResource = resourceVal
	}
	if max != "" {
		resourceVal, err := resource.ParseQuantity(max)
		if err != nil {
			return nil, err
		}
		maxResource = resourceVal
	}
	if request != "" && (min == "" || max == "") {
		resourceVal, err := resource.ParseQuantity(request)
		if err != nil {
			return nil, err
		}
		minResource = resourceVal
		maxResource = resourceVal
	}
	return &ResourceRequirementInfo{
		ResourceType: resourceType,
		MinQty:       minResource,
		MaxQty:       maxResource,
	}, nil
}

// CheckPathExists checks if a path exists or not
func CheckPathExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// path to file does exist
		return true
	}
	klog.V(4).Infof("path %s doesn't exist, skipping it", path)
	return false
}

// GetHostWithPort parses provided url and returns string formated as
// host:port even if port was not specifically specified in the origin url.
// If port is not specified, standart port corresponding to url schema is provided.
// example: for url https://example.com function will return "example.com:443"
//          for url https://example.com:8443 function will return "example:8443"
func GetHostWithPort(inputURL string) (string, error) {
	u, err := url.Parse(inputURL)
	if err != nil {
		return "", errors.Wrapf(err, "error while getting port for url %s ", inputURL)
	}

	port := u.Port()
	address := u.Host
	// if port is not specified try to detect it based on provided scheme
	if port == "" {
		portInt, err := net.LookupPort("tcp", u.Scheme)
		if err != nil {
			return "", errors.Wrapf(err, "error while getting port for url %s ", inputURL)
		}
		port = strconv.Itoa(portInt)
		address = fmt.Sprintf("%s:%s", u.Host, port)
	}
	return address, nil
}

// GetIgnoreRulesFromDirectory reads the .odoignore file, if present, and reads the rules from it
// if the .odoignore file is not found, then .gitignore is searched for the rules
// if both are not found, return empty array
// directory is the name of the directory to look into for either of the files
// rules is the array of rules (in string form)
func GetIgnoreRulesFromDirectory(directory string) ([]string, error) {
	rules := []string{".git"}
	// checking for presence of .odoignore file
	pathIgnore := filepath.Join(directory, ".odoignore")
	if _, err := os.Stat(pathIgnore); os.IsNotExist(err) || err != nil {
		// .odoignore doesn't exist
		// checking presence of .gitignore file
		pathIgnore = filepath.Join(directory, ".gitignore")
		if _, err := os.Stat(pathIgnore); os.IsNotExist(err) || err != nil {
			// both doesn't exist, return empty array
			return rules, nil
		}
	}

	file, err := os.Open(pathIgnore)
	if err != nil {
		return nil, err
	}

	defer file.Close() // #nosec G307

	scanner := bufio.NewReader(file)
	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			return rules, err
		}
		spaceTrimmedLine := strings.TrimSpace(string(line))
		if len(spaceTrimmedLine) > 0 && !strings.HasPrefix(string(line), "#") && !strings.HasPrefix(string(line), ".git") {
			rules = append(rules, string(line))
		}
	}

	return rules, nil
}

// GetAbsGlobExps converts the relative glob expressions into absolute glob expressions
// returns the absolute glob expressions
func GetAbsGlobExps(directory string, globExps []string) []string {
	absGlobExps := []string{}
	for _, globExp := range globExps {
		// for glob matching with the library
		// the relative paths in the glob expressions need to be converted to absolute paths
		absGlobExps = append(absGlobExps, filepath.Join(directory, globExp))
	}
	return absGlobExps
}

// GetSortedKeys retrieves the alphabetically-sorted keys of the specified map
func GetSortedKeys(mapping map[string]string) []string {
	keys := make([]string, len(mapping))

	i := 0
	for k := range mapping {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

// GetSplitValuesFromStr returns a slice containing the split string, using ',' as a separator
func GetSplitValuesFromStr(inputStr string) []string {
	if len(inputStr) == 0 {
		return []string{}
	}

	result := strings.Split(inputStr, ",")
	for i, value := range result {
		result[i] = strings.TrimSpace(value)
	}
	return result
}

// GetContainerPortsFromStrings generates ContainerPort values from the array of string port values
// ports is the array containing the string port values
func GetContainerPortsFromStrings(ports []string) ([]corev1.ContainerPort, error) {
	var containerPorts []corev1.ContainerPort
	for _, port := range ports {
		splits := strings.Split(port, "/")
		if len(splits) < 1 || len(splits) > 2 {
			return nil, fmt.Errorf("unable to parse the port string %s", port)
		}

		portNumberI64, err := strconv.ParseInt(splits[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid port number %s", splits[0])
		}
		portNumber := int32(portNumberI64)

		var portProto corev1.Protocol
		if len(splits) == 2 {
			switch strings.ToUpper(splits[1]) {
			case "TCP":
				portProto = corev1.ProtocolTCP
			case "UDP":
				portProto = corev1.ProtocolUDP
			default:
				return nil, fmt.Errorf("invalid port protocol %s", splits[1])
			}
		} else {
			portProto = corev1.ProtocolTCP
		}

		port := corev1.ContainerPort{
			Name:          fmt.Sprintf("%d-%s", portNumber, strings.ToLower(string(portProto))),
			ContainerPort: portNumber,
			Protocol:      portProto,
		}
		containerPorts = append(containerPorts, port)
	}
	return containerPorts, nil
}

// IsGlobExpMatch compiles strToMatch against each of the passed globExps
// Parameters:
// strToMatch : a string for matching against the rules
// globExps : a list of glob patterns to match strToMatch with
// Returns: true if there is any match else false the error (if any)
// Notes:
// Source as well as glob expression to match is changed to forward
// slashes due to supporting Windows as well as support with the
// "github.com/gobwas/glob" library that we use.
func IsGlobExpMatch(strToMatch string, globExps []string) (bool, error) {

	// Replace all backslashes with forward slashes in order for
	// glob / expression matching to work correctly with
	// the "github.com/gobwas/glob" library
	strToMatch = strings.Replace(strToMatch, "\\", "/", -1)

	for _, globExp := range globExps {

		// We replace backslashes with forward slashes for
		// glob expression / matching support
		globExp = strings.Replace(globExp, "\\", "/", -1)

		pattern, err := glob.Compile(globExp)
		if err != nil {
			return false, err
		}
		matched := pattern.Match(strToMatch)
		if matched {
			klog.V(4).Infof("ignoring path %s because of glob rule %s", strToMatch, globExp)
			return true, nil
		}
	}
	return false, nil
}

// CheckOutputFlag returns true if specified output format is supported
func CheckOutputFlag(outputFlag string) bool {
	if outputFlag == "json" || outputFlag == "" {
		return true
	}
	return false
}

// RemoveDuplicates goes through a string slice and removes all duplicates.
// Reference: https://siongui.github.io/2018/04/14/go-remove-duplicates-from-slice-or-array/
func RemoveDuplicates(s []string) []string {

	// Make a map and go through each value to see if it's a duplicate or not
	m := make(map[string]bool)
	for _, item := range s {
		if _, ok := m[item]; !ok {
			m[item] = true
		}
	}

	// Append to the unique string
	var result []string
	for item := range m {
		result = append(result, item)
	}
	return result
}

// RemoveRelativePathFromFiles removes a specified path from a list of files
func RemoveRelativePathFromFiles(files []string, path string) ([]string, error) {

	removedRelativePathFiles := []string{}
	for _, file := range files {
		rel, err := filepath.Rel(path, file)
		if err != nil {
			return []string{}, err
		}
		removedRelativePathFiles = append(removedRelativePathFiles, rel)
	}

	return removedRelativePathFiles, nil
}

// DeletePath deletes a file/directory if it exists and doesn't throw error if it doesn't exist
func DeletePath(path string) error {
	_, err := os.Stat(path)

	// reason for double negative is os.IsExist() would be blind to EMPTY FILE.
	if !os.IsNotExist(err) {
		return os.Remove(path)
	}
	return nil
}

// HTTPGetFreePort gets a free port from the system
func HTTPGetFreePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return -1, err
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	err = listener.Close()
	if err != nil {
		return -1, err
	}
	return freePort, nil
}

// IsEmpty checks to see if a directory is empty
// shamelessly taken from: https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
// this helps detect any edge cases where an empty directory is copied over
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close() // #nosec G307

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// GetRemoteFilesMarkedForDeletion returns the list of remote files marked for deletion
func GetRemoteFilesMarkedForDeletion(delSrcRelPaths []string, remoteFolder string) []string {
	var rmPaths []string
	for _, delRelPath := range delSrcRelPaths {
		// since the paths inside the container are linux oriented
		// so we convert the paths accordingly
		rmPaths = append(rmPaths, filepath.ToSlash(filepath.Join(remoteFolder, delRelPath)))
	}
	return rmPaths
}

// HTTPGetRequest gets resource contents given URL and token (if applicable)
// cacheFor determines how long the response should be cached (in minutes), 0 for no caching
func HTTPGetRequest(request HTTPRequestParams, cacheFor int) ([]byte, error) {
	// Build http request
	req, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		return nil, err
	}
	if request.Token != "" {
		bearer := "Bearer " + request.Token
		req.Header.Add("Authorization", bearer)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: ResponseHeaderTimeout,
		},
		Timeout: HTTPRequestTimeout,
	}

	klog.V(4).Infof("HTTPGetRequest: %s", req.URL.String())

	if cacheFor > 0 {
		// if there is an error during cache setup we show warning and continue without using cache
		cacheError := false
		httpCacheTime := time.Duration(cacheFor) * time.Minute

		// make sure that cache directory exists
		err = os.MkdirAll(httpCacheDir, 0750)
		if err != nil {
			cacheError = true
			klog.WarningDepth(4, "Unable to setup cache: ", err)
		}
		err = cleanHttpCache(httpCacheDir, httpCacheTime)
		if err != nil {
			cacheError = true
			klog.WarningDepth(4, "Unable to clean up cache directory: ", err)
		}

		if !cacheError {
			httpClient.Transport = httpcache.NewTransport(diskcache.New(httpCacheDir))
			klog.V(4).Infof("Response will be cached in %s for %s", httpCacheDir, httpCacheTime)
		} else {
			klog.V(4).Info("Response won't be cached.")
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Header.Get(httpcache.XFromCache) != "" {
		klog.V(4).Infof("Cached response used.")
	}

	// We have a non 1xx / 2xx status, return an error
	if (resp.StatusCode - 300) > 0 {
		return nil, errors.Errorf("failed to retrieve %s, %v: %s", request.URL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Process http response
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, err
}

// FilterIgnores applies the glob rules on the filesChanged and filesDeleted and filters them
// returns the filtered results which match any of the glob rules
func FilterIgnores(filesChanged, filesDeleted, absIgnoreRules []string) (filesChangedFiltered, filesDeletedFiltered []string) {
	for _, file := range filesChanged {
		match, err := IsGlobExpMatch(file, absIgnoreRules)
		if err != nil {
			continue
		}
		if !match {
			filesChangedFiltered = append(filesChangedFiltered, file)
		}
	}

	for _, file := range filesDeleted {
		match, err := IsGlobExpMatch(file, absIgnoreRules)
		if err != nil {
			continue
		}
		if !match {
			filesDeletedFiltered = append(filesDeletedFiltered, file)
		}
	}
	return filesChangedFiltered, filesDeletedFiltered
}

// IsValidProjectDir checks that the folder to download the project from devfile is
// either empty or only contains the devfile used.
func IsValidProjectDir(path string, devfilePath string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if len(files) > 1 {
		return errors.Errorf("Folder %s is not empty. It can only contain the devfile used.", path)
	} else if len(files) == 1 {
		file := files[0]
		if file.IsDir() {
			return errors.Errorf("Folder %s is not empty. It contains a subfolder.", path)
		}
		fileName := files[0].Name()
		devfilePath = strings.TrimPrefix(devfilePath, "./")
		if fileName != devfilePath {
			return errors.Errorf("Folder %s contains one element and it's not the devfile used.", path)
		}
	}

	return nil
}

// Converts Git ssh remote to https
func ConvertGitSSHRemoteToHTTPS(remote string) string {
	remote = strings.Replace(remote, ":", "/", 1)
	remote = strings.Replace(remote, "git@", "https://", 1)
	return remote
}

// GetAndExtractZip downloads a zip file from a URL with a http prefix or
// takes an absolute path prefixed with file:// and extracts it to a destination.
// pathToUnzip specifies the path within the zip folder to extract
func GetAndExtractZip(zipURL string, destination string, pathToUnzip string) error {
	if zipURL == "" {
		return errors.Errorf("Empty zip url: %s", zipURL)
	}
	if !strings.Contains(zipURL, ".zip") {
		return errors.Errorf("Invalid zip url: %s", zipURL)
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

		params := DownloadParams{
			Request: HTTPRequestParams{
				URL: zipURL,
			},
			Filepath: pathToZip,
		}
		err := DownloadFile(params)
		if err != nil {
			return err
		}

		defer func() {
			if err := DeletePath(pathToZip); err != nil {
				klog.Errorf("Could not delete temporary directory for zip file. Error: %s", err)
			}
		}()
	} else {
		return errors.Errorf("Invalid Zip URL: %s . Should either be prefixed with file://, http:// or https://", zipURL)
	}

	filenames, err := Unzip(pathToZip, destination, pathToUnzip)
	if err != nil {
		return err
	}

	if len(filenames) == 0 {
		return errors.New("no files were unzipped, ensure that the project repo is not empty or that sparseCheckoutDir has a valid path")
	}

	return nil
}

// Unzip will decompress a zip archive, moving specified files and folders
// within the zip file (parameter 1) to an output directory (parameter 2)
// Source: https://golangcode.com/unzip-files-in-go/
// pathToUnzip (parameter 3) is the path within the zip folder to extract
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

		// if sparseCheckoutDir has a pattern
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

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, ModeReadWriteFile)
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

// DownloadFileWithCache downloads the file to the filepath given URL and token (if applicable)
// cacheFor determines how long the response should be cached (in minutes), 0 for no caching
func DownloadFileWithCache(params DownloadParams, cacheFor int) error {
	// Get the data
	data, err := HTTPGetRequest(params.Request, cacheFor)
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(params.Filepath)
	if err != nil {
		return err
	}
	defer out.Close() // #nosec G307

	// Write the data to file
	_, err = out.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// DownloadFile downloads the file to the filepath given URL and token (if applicable)
func DownloadFile(params DownloadParams) error {
	return DownloadFileWithCache(params, 0)
}

// DownloadFileInMemory uses the url to download the file and return bytes
func DownloadFileInMemory(url string) ([]byte, error) {
	var httpClient = &http.Client{Transport: &http.Transport{
		ResponseHeaderTimeout: ResponseHeaderTimeout,
	}, Timeout: HTTPRequestTimeout}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	// We have a non 1xx / 2xx status, return an error
	if (resp.StatusCode - 300) > 0 {
		return nil, errors.Errorf("failed to retrieve %s, %v: %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// ValidateK8sResourceName sanitizes kubernetes resource name with the following requirements:
// - Contain at most 63 characters
// - Contain only lowercase alphanumeric characters or ‘-’
// - Start with an alphanumeric character
// - End with an alphanumeric character
// - Must not contain all numeric values
func ValidateK8sResourceName(key string, value string) error {
	requirements := `
- Contain at most 63 characters
- Contain only lowercase alphanumeric characters or ‘-’
- Start with an alphanumeric character
- End with an alphanumeric character
- Must not contain all numeric values
	`
	err1 := kvalidation.IsDNS1123Label(value)
	_, err2 := strconv.ParseFloat(value, 64)

	if err1 != nil || err2 == nil {
		return errors.Errorf("%s \"%s\" is not valid, %s should conform the following requirements: %s", key, value, key, requirements)
	}

	return nil
}

// CheckKubeConfigExist checks for existence of kubeconfig
func CheckKubeConfigExist() bool {

	var kubeconfig string

	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	} else {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
			klog.V(4).Infof("using default kubeconfig path %s", kubeconfig)
		} else {
			klog.V(4).Infof("no KUBECONFIG provided and cannot fallback to default")
			return false
		}
	}

	if CheckPathExists(kubeconfig) {
		return true
	}

	return false
}

// ValidateURL validates the URL
func ValidateURL(sourceURL string) error {
	u, err := url.Parse(sourceURL)
	if err != nil {
		return err
	}

	if len(u.Host) == 0 || len(u.Scheme) == 0 {
		return errors.New("URL is invalid")
	}

	return nil
}

// ValidateFile validates the file
func ValidateFile(filePath string) error {
	// Check if the file path exist
	file, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if file.IsDir() {
		return errors.Errorf("%s exists but it's not a file", filePath)
	}

	return nil
}

// CopyFile copies file from source path to destination path
func CopyFile(srcPath string, dstPath string, info os.FileInfo) error {
	// In order to avoid file overriding issue, do nothing if source path is equal to destination path
	if PathEqual(srcPath, dstPath) {
		return nil
	}
	// Check if the source file path exists
	err := ValidateFile(srcPath)
	if err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close() // #nosec G307

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close() // #nosec G307

	// Ensure destination file has the same file mode with source file
	err = os.Chmod(dstFile.Name(), info.Mode())
	if err != nil {
		return err
	}

	// Copy file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// PathEqual compare the paths to determine if they are equal
func PathEqual(firstPath string, secondPath string) bool {
	firstAbsPath, _ := GetAbsPath(firstPath)
	secondAbsPath, _ := GetAbsPath(secondPath)
	return firstAbsPath == secondAbsPath
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

// AddFileToIgnoreFile adds a file to the gitignore file. It only does that if the file doesn't exist
func AddFileToIgnoreFile(gitIgnoreFile, filename string) error {
	return addFileToIgnoreFile(gitIgnoreFile, filename, filesystem.DefaultFs{})
}

func addFileToIgnoreFile(gitIgnoreFile, filename string, fs filesystem.Filesystem) error {
	var data []byte
	file, err := fs.OpenFile(gitIgnoreFile, os.O_APPEND|os.O_RDWR, ModeReadWriteFile)
	if err != nil {
		return errors.Wrap(err, "failed to open .gitignore file")
	}
	defer file.Close()

	if data, err = fs.ReadFile(gitIgnoreFile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed reading data from %v file", gitIgnoreFile))
	}
	// check whether .odo/odo-file-index.json is already in the .gitignore file
	if !strings.Contains(string(data), filename) {
		if _, err := file.WriteString("\n" + filename); err != nil {
			return errors.Wrapf(err, "failed to add %v to %v file", filepath.Base(filename), gitIgnoreFile)
		}
	}
	return nil
}

// DisplayLog displays logs to user stdout with some color formatting
func DisplayLog(followLog bool, rd io.ReadCloser, compName string) (err error) {

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

		if _, err = io.Copy(os.Stdout, rd); err != nil {
			return errors.Wrapf(err, "error followLoging logs for %s", compName)
		}

	} else {

		// Copy to buffer (we aren't going to be followLoging the logs..)
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rd)
		if err != nil {
			return errors.Wrapf(err, "unable to copy followLog to buffer")
		}

		// Copy to stdout
		if _, err = io.Copy(os.Stdout, buf); err != nil {
			return errors.Wrapf(err, "error copying logs to stdout")
		}

	}
	return

}
