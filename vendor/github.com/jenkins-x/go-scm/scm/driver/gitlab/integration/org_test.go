// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

//
// organization sub-tests
//

func testOrgs(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testOrgFind(client))
	}
}

func testOrgFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Organizations.Find(context.Background(), "diaspora")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Organization", testOrg(result))
	}
}

//
// struct sub-tests
//

func testOrg(organization *scm.Organization) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		if got, want := organization.Name, "diaspora"; got != want {
			t.Errorf("Want organization Name %q, got %q", want, got)
		}
		if got, want := organization.Avatar, ""; got != want {
			t.Errorf("Want organization Avatar %q, got %q", want, got)
		}
	}
}
