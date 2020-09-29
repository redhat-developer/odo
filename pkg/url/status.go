package url

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"

	"k8s.io/klog"
)

const (
	// URLFailureWaitTime is how long to wait on error from URL connection
	URLFailureWaitTime = time.Duration(5) * time.Second
)

// StartURLHttpRequestStatusWatchForK8S begins testing URLs for responses, outputting the result to console
func StartURLHttpRequestStatusWatchForK8S(occlient *occlient.Client, client *kclient.Client, envInfo *envinfo.EnvSpecificInfo, loggingClient machineoutput.MachineEventLoggingClient, containerComponents []common.DevfileComponent) {

	// This is a non-blocking function so that other status watchers may start as needed
	go func() {

		var urlList []statusURL

		for {
			var err error
			urlList, err = getURLsForKubernetes(occlient, client, envInfo, true, containerComponents)

			if err == nil {
				// Success!
				break
			} else {
				// Try again in a few seconds...
				klog.V(4).Infof("Unable to get URLs: %v", err)
				time.Sleep(URLFailureWaitTime)
			}
		}

		if len(urlList) == 0 {
			return
		}

		singleEntry := [][]statusURL{urlList}

		startURLTester(singleEntry, loggingClient)

	}()
}

// StartURLHttpRequestStatusWatchForDocker begins testing URLs for responses, outputting the result to console
func StartURLHttpRequestStatusWatchForDocker(envInfo *envinfo.EnvSpecificInfo, loggingClient machineoutput.MachineEventLoggingClient) {

	// This is a non-blocking function so that other status watchers may start as needed
	go func() {

		// Since Docker URLs do not have a protocol (http/https) associated with them, this URL list will contain URLs for
		// both protocols.
		var urlList [][]statusURL

		for {
			var err error
			urlList, err = getURLsForDocker(envInfo)

			if err == nil {
				// Success!
				break
			} else {
				// Try again in a few seconds...
				klog.V(4).Infof("Unable to get Docker URLs: %v", err)
				time.Sleep(URLFailureWaitTime)
			}
		}

		startURLTester(urlList, loggingClient)

	}()
}

// startURLTester kicks off a new goroutine for each set of URLs to test
func startURLTester(urlsToTest [][]statusURL, loggingClient machineoutput.MachineEventLoggingClient) {

	for _, urlToTest := range urlsToTest {
		startURLTestGoRoutine(urlToTest, URLFailureWaitTime, loggingClient)
	}
}

func getURLsForKubernetes(oclient *occlient.Client, client *kclient.Client, envInfo *envinfo.EnvSpecificInfo, ignoreUnpushed bool, containerComponents []common.DevfileComponent) ([]statusURL, error) {
	componentName := envInfo.GetName()

	routesSupported := false

	if oclient != nil {
		var err error

		if routesSupported, err = oclient.IsRouteSupported(); err != nil {
			// Fallback to Kubernetes client on error
			routesSupported = false
			oclient = nil
		}

	}

	urls, err := ListIngressAndRoute(oclient, envInfo, containerComponents, componentName, routesSupported)

	if err != nil {
		return nil, err
	}
	urlList := []statusURL{}

	for _, u := range urls.Items {

		// Ignore unpushed URLs, they necessarily are unreachable
		if u.Status.State != StateTypePushed && ignoreUnpushed {
			continue
		}

		var properURL, protocol string

		if u.Spec.Kind != envinfo.ROUTE {
			protocol = GetProtocol(routev1.Route{}, ConvertIngressURLToIngress(u, componentName))
			properURL = GetURLString(protocol, "", u.Spec.Host, false)
		} else {
			protocol = u.Spec.Protocol
			properURL = GetURLString(protocol, u.Spec.Host, "", false)
		}

		statusURLVal := statusURL{
			name:   u.Name,
			url:    properURL,
			kind:   string(u.Spec.Kind),
			port:   u.Spec.Port,
			secure: protocol == "https",
		}

		urlList = append(urlList, statusURLVal)

	}

	return urlList, nil
}

type statusURL struct {
	name   string
	url    string
	port   int
	secure bool
	kind   string
}

func getURLsForDocker(envInfo *envinfo.EnvSpecificInfo) ([][]statusURL, error) {

	componentName := envInfo.GetName()
	client, err := lclient.New()
	if err != nil {
		return nil, err
	}

	urls, err := ListDockerURL(client, componentName, envInfo)
	if err != nil {
		return nil, err
	}

	urlList := [][]statusURL{}

	for _, u := range urls.Items {
		var urlString string
		if u.Status.State == StateTypeNotPushed {
			continue
		}

		urlPair := []statusURL{}

		// Docker URLs (unlike K8s URLs) do not have a protocol assigned to them, so we must add both HTTP and HTTPS

		// Add http://
		urlString = fmt.Sprintf("http://%s:%s", u.Spec.Host, strconv.Itoa(u.Spec.ExternalPort))

		statusURLVal := statusURL{
			name:   u.Name,
			url:    urlString,
			kind:   "docker",
			port:   u.Spec.ExternalPort,
			secure: false,
		}

		urlPair = append(urlPair, statusURLVal)

		// Add https://
		urlString = fmt.Sprintf("https://%s:%s", u.Spec.Host, strconv.Itoa(u.Spec.ExternalPort))

		statusURLVal = statusURL{
			name:   u.Name,
			url:    urlString,
			kind:   "docker",
			port:   u.Spec.ExternalPort,
			secure: true,
		}
		urlPair = append(urlPair, statusURLVal)

		urlList = append(urlList, urlPair)
	}

	return urlList, nil
}

// startURLTestGoRoutine tests one or more urls ('urls' param); if at least one of them is successful, a success is reported.
// If a success was previously reported, additional successes will not be reported (until at least one failure occurs)
// Likewise if a failure was previously reported.
func startURLTestGoRoutine(urls []statusURL, delayBetweenRequests time.Duration, loggingClient machineoutput.MachineEventLoggingClient) {

	go func() {

		var previousResult *bool = nil

		for {

			successfulMatch := (*statusURL)(nil)
			atLeastOneSuccess := false

			for _, currURL := range urls {
				anyResponseReceived, err := testURL(currURL.url)

				if err != nil {
					klog.V(4).Infof("Error on connecting to URL '%s' %v", currURL.url, err)
				}

				if anyResponseReceived {
					match := currURL
					successfulMatch = &match
					atLeastOneSuccess = true
					break
				}
			}

			// If this is the first time we have seen a result for this URL, OR the result has changed from last time
			if previousResult == nil || *previousResult != atLeastOneSuccess {

				if successfulMatch != nil {
					// At least one of the URLs was reachable, so report success for it
					loggingClient.URLReachable((*successfulMatch).name, (*successfulMatch).url, (*successfulMatch).port, (*successfulMatch).secure, (*successfulMatch).kind, true, machineoutput.TimestampNow())
				} else {
					// Otherwise report failure for all URLs
					for _, currURL := range urls {
						loggingClient.URLReachable(currURL.name, currURL.url, currURL.port, currURL.secure, currURL.kind, atLeastOneSuccess, machineoutput.TimestampNow())
					}
				}
			}

			previousResult = &atLeastOneSuccess

			time.Sleep(delayBetweenRequests)
		}
	}()
}

// testURL tests a single URL, returning true if ANY response was received, or false otherwise (plus an error if applicable)
func testURL(url string) (bool, error) {

	// Suppress 'G402 (CWE-295): TLS InsecureSkipVerify set true': since we are not using the contents of the HTTP response,
	// the fact that it can be MITM-ed is irrelevant.
	/* #nosec */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil || resp == nil {
		errMsg := "Get request failed for " + url + " , with no response code."
		if err != nil {
			return false, errors.Wrapf(err, errMsg)
		}
		return false, errors.New(errMsg)
	}

	defer resp.Body.Close()

	// Any response code (including 5XX, 4XX) is considered a success; we only use this to determine if the process
	// is responding to requests.
	klog.V(4).Infof("Get request succeeded for '%s', with response code %d", url, resp.StatusCode)

	return true, nil

}
