// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

type userService struct {
	client *wrapper
}

func (s *userService) Find(ctx context.Context) (*scm.User, *scm.Response, error) {
	out := new(user)
	res, err := s.client.do(ctx, "GET", "api/v1/user", nil, out)
	return convertUser(out), res, err
}

func (s *userService) FindLogin(ctx context.Context, login string) (*scm.User, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/users/%s", login)
	out := new(user)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertUser(out), res, err
}

func (s *userService) FindEmail(ctx context.Context) (string, *scm.Response, error) {
	user, res, err := s.Find(ctx)
	return user.Email, res, err
}

//
// native data structures
//

type user struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Username string `json:"username"`
	Fullname string `json:"full_name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
}

//
// native data structure conversion
//

func convertUser(src *user) *scm.User {
	return &scm.User{
		Login:  userLogin(src),
		Avatar: src.Avatar,
		Email:  src.Email,
		Name:   src.Fullname,
	}
}

func userLogin(src *user) string {
	if src.Username != "" {
		return src.Username
	}
	return src.Login
}
