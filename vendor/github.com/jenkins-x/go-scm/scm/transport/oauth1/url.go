// Copyright (c) 2015 Dalton Hubble. All rights reserved.
// Copyrights licensed under the MIT License.

package oauth1

import (
	"fmt"
	"net/http"
	"strings"
)

// baseURI returns the base string URI of a request
// according to RFC 5849 3.4.1.2.
func baseURI(r *http.Request) string {
	scheme := strings.ToLower(r.URL.Scheme)
	host := strings.ToLower(r.URL.Host)
	if hostPort := strings.Split(host, ":"); len(hostPort) == 2 && (hostPort[1] == "80" || hostPort[1] == "443") {
		host = hostPort[0]
	}
	path := r.URL.EscapedPath()
	return fmt.Sprintf("%v://%v%v", scheme, host, path)
}
