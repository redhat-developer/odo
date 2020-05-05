// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

type organizationService struct {
	client *wrapper
}

func (s *organizationService) ListTeams(ctx context.Context, org string, ops scm.ListOptions) ([]*scm.Team, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *organizationService) ListTeamMembers(ctx context.Context, id int, role string, ops scm.ListOptions) ([]*scm.TeamMember, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *organizationService) ListOrgMembers(ctx context.Context, org string, ops scm.ListOptions) ([]*scm.TeamMember, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *organizationService) IsMember(ctx context.Context, org string, user string) (bool, *scm.Response, error) {
	opts := scm.ListOptions{
		Size: 1000,
	}
	path := fmt.Sprintf("rest/api/1.0/projects/%s/permissions/users?%s", org, encodeListOptions(opts))
	out := new(participants)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	if !out.pagination.LastPage.Bool {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	for _, participant := range out.Values {
		if participant.User.Name == user || participant.User.Slug == user {
			return true, res, err
		}
	}
	return false, res, err
}

func (s *organizationService) IsAdmin(ctx context.Context, org string, user string) (bool, *scm.Response, error) {
	return false, nil, scm.ErrNotSupported
}

func (s *organizationService) Find(ctx context.Context, name string) (*scm.Organization, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *organizationService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Organization, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}
