// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogs

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

type organizationService struct {
	client *wrapper
}

func (s *organizationService) IsMember(ctx context.Context, org string, user string) (bool, *scm.Response, error) {
	return false, nil, scm.ErrNotSupported
}

func (s *organizationService) IsAdmin(ctx context.Context, org string, user string) (bool, *scm.Response, error) {
	return false, nil, scm.ErrNotSupported
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

func (s *organizationService) Find(ctx context.Context, name string) (*scm.Organization, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/orgs/%s", name)
	out := new(org)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertOrg(out), res, err
}

func (s *organizationService) List(ctx context.Context, _ scm.ListOptions) ([]*scm.Organization, *scm.Response, error) {
	var out []*org
	res, err := s.client.do(ctx, "GET", "api/v1/user/orgs", nil, &out)
	return convertOrgList(out), res, err
}

//
// native data structures
//

type org struct {
	Name   string `json:"username"`
	Avatar string `json:"avatar_url"`
}

//
// native data structure conversion
//

func convertOrgList(from []*org) []*scm.Organization {
	to := []*scm.Organization{}
	for _, v := range from {
		to = append(to, convertOrg(v))
	}
	return to
}

func convertOrg(from *org) *scm.Organization {
	return &scm.Organization{
		Name:   from.Name,
		Avatar: from.Avatar,
	}
}
