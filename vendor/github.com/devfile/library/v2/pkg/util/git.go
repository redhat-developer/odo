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

package util

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitHubHost    string = "github.com"
	RawGitHubHost string = "raw.githubusercontent.com"
	GitLabHost    string = "gitlab.com"
	BitbucketHost string = "bitbucket.org"
)

type GitUrl struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Revision string // branch name, tag name, or commit id
	Path     string // path to a directory or file in the repo
	Token    string // authenticates private repo actions for parent devfiles
	IsFile   bool   // defines if the URL points to a file in the repo
}

// NewGitURL NewGitUrl creates a GitUrl from a string url and token.  Will eventually replace NewGitUrlWithURL
func NewGitURL(url string, token string) (*GitUrl, error) {
	gitUrl, err := ParseGitUrl(url)
	if err != nil {
		return &gitUrl, err
	}
	gitUrl.Token = token
	return &gitUrl, nil
}

// NewGitUrlWithURL NewGitUrl creates a GitUrl from a string url
func NewGitUrlWithURL(url string) (GitUrl, error) {
	gitUrl, err := ParseGitUrl(url)
	if err != nil {
		return gitUrl, err
	}
	return gitUrl, nil
}

// ParseGitUrl extracts information from a support git url
// Only supports git repositories hosted on GitHub, GitLab, and Bitbucket
func ParseGitUrl(fullUrl string) (GitUrl, error) {
	var g GitUrl
	err := ValidateURL(fullUrl)
	if err != nil {
		return g, err
	}

	parsedUrl, err := url.Parse(fullUrl)
	if err != nil {
		return g, err
	}

	if len(parsedUrl.Path) == 0 {
		return g, fmt.Errorf("url path should not be empty")
	}

	if parsedUrl.Host == RawGitHubHost || parsedUrl.Host == GitHubHost {
		err = g.parseGitHubUrl(parsedUrl)
	} else if parsedUrl.Host == GitLabHost {
		err = g.parseGitLabUrl(parsedUrl)
	} else if parsedUrl.Host == BitbucketHost {
		err = g.parseBitbucketUrl(parsedUrl)
	} else {
		err = fmt.Errorf("url host should be a valid GitHub, GitLab, or Bitbucket host; received: %s", parsedUrl.Host)
	}

	return g, err
}

func (g *GitUrl) GetToken() string {
	return g.Token
}

type CommandType string

const (
	GitCommand        CommandType = "git"
	unsupportedCmdMsg             = "Unsupported command \"%s\" "
)

// Execute is exposed as a global variable for the purpose of running mock tests
// only "git" is supported
/* #nosec G204 -- used internally to execute various git actions and eventual cleanup of artifacts.  Calling methods validate user input to ensure commands are used appropriately */
var execute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		c := exec.Command(string(cmd), args...)
		c.Dir = baseDir
		output, err := c.CombinedOutput()
		return output, err
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

func (g *GitUrl) CloneGitRepo(destDir string) error {
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := g.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if g.GetToken() == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, host, g.Owner, g.Repo)
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", g.Protocol, g.GetToken(), host, g.Owner, g.Repo)
		if g.Host == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", g.Protocol, g.GetToken(), host, g.Owner, g.Repo)
		}
	}

	_, err := execute(destDir, "git", "clone", repoUrl, destDir)

	if err != nil {
		if g.GetToken() == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private. error: %v", err)
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct. error: %v", err)
		}
	}

	if g.Revision != "" {
		_, err := execute(destDir, "git", "checkout", g.Revision)
		if err != nil {
			err = os.RemoveAll(destDir)
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to switch repo to revision. repo dir: %v, revision: %v, error: %v", destDir, g.Revision, err)
		}
	}

	return nil
}

func (g *GitUrl) parseGitHubUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host

	if g.Host == RawGitHubHost {
		g.IsFile = true
		// raw GitHub urls don't contain "blob" or "tree"
		// https://raw.githubusercontent.com/devfile/library/main/devfile.yaml -> [devfile library main devfile.yaml]
		splitUrl = strings.SplitN(url.Path[1:], "/", 4)
		if len(splitUrl) == 4 {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]
			g.Revision = splitUrl[2]
			g.Path = splitUrl[3]
		} else {
			// raw GitHub urls have to be a file
			err = fmt.Errorf("raw url path should contain <owner>/<repo>/<branch>/<path/to/file>, received: %s", url.Path[1:])
		}
		return err
	}

	if g.Host == GitHubHost {
		// https://github.com/devfile/library/blob/main/devfile.yaml -> [devfile library blob main devfile.yaml]
		splitUrl = strings.SplitN(url.Path[1:], "/", 5)
		if len(splitUrl) < 2 {
			err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
		} else {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]

			// url doesn't contain a path to a directory or file
			if len(splitUrl) == 2 {
				return nil
			}

			switch splitUrl[2] {
			case "tree":
				g.IsFile = false
			case "blob":
				g.IsFile = true
			default:
				return fmt.Errorf("url path to directory or file should contain 'tree' or 'blob'")
			}

			// url has a path to a file or directory
			if len(splitUrl) == 5 {
				g.Revision = splitUrl[3]
				g.Path = splitUrl[4]
			} else if !g.IsFile && len(splitUrl) == 4 {
				g.Revision = splitUrl[3]
			} else {
				err = fmt.Errorf("url path should contain <owner>/<repo>/<tree or blob>/<branch>/<path/to/file/or/directory>, received: %s", url.Path[1:])
			}
		}
	}

	return err
}

func (g *GitUrl) parseGitLabUrl(url *url.URL) error {
	var splitFile, splitOrg []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

	// GitLab urls contain a '-' separating the root of the repo
	// and the path to a file or directory
	split := strings.Split(url.Path[1:], "/-/")

	splitOrg = strings.SplitN(split[0], "/", 2)
	if len(splitOrg) < 2 {
		return fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else {
		g.Owner = splitOrg[0]
		g.Repo = splitOrg[1]
	}

	// url doesn't contain a path to a directory or file
	if len(split) == 1 {
		return nil
	}

	// url may contain a path to a directory or file
	if len(split) == 2 {
		splitFile = strings.SplitN(split[1], "/", 3)
	}

	if len(splitFile) == 3 {
		if splitFile[0] == "blob" || splitFile[0] == "tree" || splitFile[0] == "raw" {
			g.Revision = splitFile[1]
			g.Path = splitFile[2]
			ext := filepath.Ext(g.Path)
			if ext != "" {
				g.IsFile = true
			}
		} else {
			err = fmt.Errorf("url path should contain 'blob' or 'tree' or 'raw', received: %s", url.Path[1:])
		}
	} else {
		return fmt.Errorf("url path to directory or file should contain <blob or tree or raw>/<branch>/<path/to/file/or/directory>, received: %s", url.Path[1:])
	}

	return err
}

func (g *GitUrl) parseBitbucketUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

	splitUrl = strings.SplitN(url.Path[1:], "/", 5)
	if len(splitUrl) < 2 {
		err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else if len(splitUrl) == 2 {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
	} else {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
		if len(splitUrl) == 5 {
			if splitUrl[2] == "raw" || splitUrl[2] == "src" {
				g.Revision = splitUrl[3]
				g.Path = splitUrl[4]
				ext := filepath.Ext(g.Path)
				if ext != "" {
					g.IsFile = true
				}
			} else {
				err = fmt.Errorf("url path should contain 'raw' or 'src', received: %s", url.Path[1:])
			}
		} else {
			err = fmt.Errorf("url path should contain path to directory or file, received: %s", url.Path[1:])
		}
	}

	return err
}

// SetToken validates the token with a get request to the repo before setting the token
// Defaults token to empty on failure.
// Deprecated.  Avoid using since this will cause rate limiting issues
func (g *GitUrl) SetToken(token string, httpTimeout *int) error {
	err := g.validateToken(HTTPRequestParams{Token: token, Timeout: httpTimeout})
	if err != nil {
		g.Token = ""
		return fmt.Errorf("failed to set token. error: %v", err)
	}
	g.Token = token
	return nil
}

// IsPublic checks if the GitUrl is public with a get request to the repo using an empty token
// Returns true if the request succeeds
// Deprecated.  Avoid using since this will cause rate limiting issues
func (g *GitUrl) IsPublic(httpTimeout *int) bool {
	err := g.validateToken(HTTPRequestParams{Token: "", Timeout: httpTimeout})
	if err != nil {
		return false
	}
	return true
}

// validateToken makes a http get request to the repo with the GitUrl token
// Returns an error if the get request fails
func (g *GitUrl) validateToken(params HTTPRequestParams) error {
	var apiUrl string

	switch g.Host {
	case GitHubHost, RawGitHubHost:
		apiUrl = fmt.Sprintf("https://api.github.com/repos/%s/%s", g.Owner, g.Repo)
	case GitLabHost:
		apiUrl = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s", g.Owner, g.Repo)
	case BitbucketHost:
		apiUrl = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", g.Owner, g.Repo)
	default:
		apiUrl = fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, g.Host, g.Owner, g.Repo)
	}

	params.URL = apiUrl
	res, err := HTTPGetRequest(params, 0)
	if len(res) == 0 || err != nil {
		return err
	}

	return nil
}

// GitRawFileAPI returns the endpoint for the git providers raw file
func (g *GitUrl) GitRawFileAPI() string {
	var apiRawFile string

	switch g.Host {
	case GitHubHost, RawGitHubHost:
		apiRawFile = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", g.Owner, g.Repo, g.Revision, g.Path)
	case GitLabHost:
		apiRawFile = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/repository/files/%s/raw?ref=%s", g.Owner, g.Repo, g.Path, g.Revision)
	case BitbucketHost:
		apiRawFile = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/%s", g.Owner, g.Repo, g.Revision, g.Path)
	}

	return apiRawFile
}

// IsGitProviderRepo checks if the url matches a repo from a supported git provider
func (g *GitUrl) IsGitProviderRepo() bool {
	switch g.Host {
	case GitHubHost, RawGitHubHost, GitLabHost, BitbucketHost:
		return true
	default:
		return false
	}
}
