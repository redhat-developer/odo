// Copyright (c) 2015 Dalton Hubble. All rights reserved.
// Copyrights licensed under the MIT License.

package oauth1

import (
	"fmt"
	"sort"
)

// sortParameters sorts parameters by key and returns a
// slice of formatted key value pairs.
func sortParameters(params map[string]string, format string) []string {
	// sort by key
	keys := make([]string, len(params))
	i := 0
	for key := range params {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	// parameter join
	pairs := make([]string, len(params))
	for i, key := range keys {
		pairs[i] = fmt.Sprintf(format, key, params[key])
	}
	return pairs
}
