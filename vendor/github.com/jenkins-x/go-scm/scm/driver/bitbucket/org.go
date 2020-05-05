// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

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
	path := fmt.Sprintf("2.0/teams/%s", name)
	out := new(organization)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertOrganization(out), res, err
}

func (s *organizationService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Organization, *scm.Response, error) {
	path := fmt.Sprintf("2.0/teams?%s", encodeListRoleOptions(opts))
	out := new(organizationList)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	copyPagination(out.pagination, res)
	return convertOrganizationList(out), res, err
}

func convertOrganizationList(from *organizationList) []*scm.Organization {
	to := []*scm.Organization{}
	for _, v := range from.Values {
		to = append(to, convertOrganization(v))
	}
	return to
}

type organizationList struct {
	pagination
	Values []*organization `json:"values"`
}

type organization struct {
	Login string `json:"username"`
}

func convertOrganization(from *organization) *scm.Organization {
	return &scm.Organization{
		Name:   from.Login,
		Avatar: fmt.Sprintf("https://bitbucket.org/account/%s/avatar/32/", from.Login),
	}
}
