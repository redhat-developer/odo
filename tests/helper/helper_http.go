package helper

import (
	"crypto/tls"
	"fmt"
	"github.com/redhat-developer/odo/pkg/util"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

// HttpWaitForWithStatus periodically (every interval) calls GET to given url
// ends when result response contains match string and status code, or after the maxRetry
func HttpWaitForWithStatus(url string, match string, maxRetry int, interval int, expectedCode int) {
	fmt.Fprintf(GinkgoWriter, "Checking %s, for %s\n", url, match)

	var body []byte

	for i := 0; i < maxRetry; i++ {
		fmt.Fprintf(GinkgoWriter, "try %d of %d\n", i, maxRetry)

		// #nosec
		// gosec:G107, G402 -> This is safe since it's just used for testing.
		transporter := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: transporter}
		resp, err := client.Get(url)
		if err != nil {
			// we log the error and sleep again because this could mean the component is not up yet
			fmt.Fprintln(GinkgoWriter, "error while requesting:", err.Error())
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == expectedCode {
			body, _ = io.ReadAll(resp.Body)
			if strings.Contains(string(body), match) {
				return
			}

		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	fmt.Fprintf(GinkgoWriter, "Last output from %s: %s\n", url, string(body))
	Fail(fmt.Sprintf("Failed after %d retries. Content in %s doesn't include '%s'.", maxRetry, url, match))
}

// GetRandomFreePort returns a random free port(string) between 1024 and 65535, or within a given portRange if provided
// WARN: If length of portRange is anything other than 2 and first index is greater-than-or-equal-to second index, it will use the default range.
func GetRandomFreePort(portRange ...int) string {
	rand.Seed(time.Now().UnixNano())
	max := 65535
	min := 1024
	var (
		startPort = rand.Intn(max-min) + min // #nosec  // cannot use crypto/rand library here
		endPort   = max
	)
	// WARN: If length of portRange is anything other than 2 and first index is gte second index, it will be ignored
	if len(portRange) == 2 && portRange[0] < portRange[1] {
		startPort = portRange[0]
		endPort = portRange[1]
	}
	freePort, err := util.NextFreePort(startPort, endPort, nil)
	if err != nil {
		Fail("failed to obtain a free port")
	}
	return strconv.Itoa(freePort)

}
