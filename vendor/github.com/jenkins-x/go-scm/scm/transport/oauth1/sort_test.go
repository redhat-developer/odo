// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSortParameters(t *testing.T) {
	params := map[string]string{
		"page":                   "1",
		"per_page":               "25",
		"oauth_version":          "1.0",
		"oauth_signature_method": "RSA-SHA1",
		"oauth_consumer_key":     "12345",
	}
	want := []string{
		"oauth_consumer_key=12345",
		"oauth_signature_method=RSA-SHA1",
		"oauth_version=1.0",
		"page=1",
		"per_page=25",
	}
	got := sortParameters(params, "%s=%s")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}
