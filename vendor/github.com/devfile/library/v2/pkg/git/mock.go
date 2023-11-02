//
// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type MockGitUrl struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Revision string // branch name, tag name, or commit id
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

func (m *MockGitUrl) GetToken() string {
	return m.token
}

var mockExecute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		if len(args) > 0 && args[0] == "clone" {
			u, _ := url.Parse(args[1])
			password, hasPassword := u.User.Password()

			resourceFile, err := os.Create(filepath.Clean(baseDir) + "/resource.file")
			if err != nil {
				return nil, fmt.Errorf("failed to create test resource: %v", err)
			}

			// private repository
			if hasPassword {
				switch password {
				case "valid-token":
					_, err := resourceFile.WriteString("private repo\n")
					if err != nil {
						return nil, fmt.Errorf("failed to write to test resource: %v", err)
					}
					return []byte(""), nil
				default:
					return []byte(""), fmt.Errorf("not a valid token")
				}
			}

			_, err = resourceFile.WriteString("public repo\n")
			if err != nil {
				return nil, fmt.Errorf("failed to write to test resource: %v", err)
			}
			return []byte(""), nil
		}

		if len(args) > 0 && args[0] == "checkout" {
			revision := args[1]
			if revision != "invalid-revision" {
				resourceFile, err := os.OpenFile(filepath.Clean(baseDir)+"/resource.file", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					return nil, fmt.Errorf("failed to open test resource: %v", err)
				}
				_, err = resourceFile.WriteString("git switched")
				if err != nil {
					return nil, fmt.Errorf("failed to write to test resource: %v", err)
				}
				return []byte("git switched to revision"), nil
			}
			return []byte(""), fmt.Errorf("failed to switch revision")
		}
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

func (m *MockGitUrl) CloneGitRepo(destDir string) error {
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := m.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if m.GetToken() == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", m.Protocol, host, m.Owner, m.Repo)
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", m.Protocol, m.GetToken(), host, m.Owner, m.Repo)
		if m.Host == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", m.Protocol, m.GetToken(), host, m.Owner, m.Repo)
		}
	}

	_, err := mockExecute(destDir, "git", "clone", repoUrl, destDir)

	if err != nil {
		if m.GetToken() == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private")
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct")
		}
	}

	if m.Revision != "" {
		_, err := mockExecute(destDir, "git", "checkout", m.Revision)
		if err != nil {
			return fmt.Errorf("failed to switch repo to revision. repo dir: %v, revision: %v", destDir, m.Revision)
		}
	}

	return nil
}

func (m *MockGitUrl) SetToken(token string) error {
	m.token = token
	return nil
}

func (m *MockGitUrl) IsGitProviderRepo() bool {
	switch m.Host {
	case GitHubHost, RawGitHubHost, GitLabHost, BitbucketHost:
		return true
	default:
		return false
	}
}
