package helper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
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

// GetCustomStartPort returns a port that can be used as starting value for custom port mapping.
// Because of the way Ginkgo runs specs in parallel (by isolating them in different processes),
// this function needs to be called in a Before* node or test spec.
// It returns a starting value that aims at minimizing the probability of collisions.
// Callers can then safely increment the returned value in their specs if needed.
func GetCustomStartPort() int {
	return 30000 + 100*GinkgoParallelProcess()
}
