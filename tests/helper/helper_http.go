package helper

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			body, _ = ioutil.ReadAll(resp.Body)
			if strings.Contains(string(body), match) {
				return
			}

		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
	fmt.Fprintf(GinkgoWriter, "Last output from %s: %s\n", url, string(body))
	Fail(fmt.Sprintf("Failed after %d retries. Content in %s doesn't include '%s'.", maxRetry, url, match))
}

// HttpWaitFor periodically (every interval) calls GET to given url
// ends when a 200 HTTP result response contains match string, or after the maxRetry
func HttpWaitFor(url string, match string, maxRetry int, interval int) {
	HttpWaitForWithStatus(url, match, maxRetry, interval, 200)
}

// HttpFileServer starts a http server with a file handler on the free port provided
// the file handler uses the location provided for serving the requests
func HttpFileServer(port int, location string) *http.Server {
	addressLook := "localhost:" + strconv.Itoa(port)
	fileHandler := http.FileServer(http.Dir(location))

	server := &http.Server{Addr: addressLook, Handler: fileHandler}
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			Expect(err).To(BeNil())
		}
	}()
	return server
}
