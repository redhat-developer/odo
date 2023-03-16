package registry

import (
	"errors"
	"fmt"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"os"
	"path/filepath"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/util"
)

const (
	RegistryUser = "default"
)

func checkoutProject(subDir, zipURL, path, starterToken string, fsys filesystem.Filesystem) error {

	if subDir == "" {
		subDir = "/"
	}
	err := util.GetAndExtractZip(zipURL, path, subDir, starterToken, fsys)
	if err != nil {
		return fmt.Errorf("failed to download and extract project zip folder: %w", err)
	}
	return nil
}

// DownloadStarterProject downloads a starter project referenced in devfile
// This will first remove the content of the contextDir
func DownloadStarterProject(fs filesystem.Filesystem, starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error {
	var path string
	var err error
	// Retrieve the working directory in order to clone correctly
	if contextDir == "" {
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get the current working directory: %w", err)
		}
	} else {
		path = contextDir
	}

	// We will check to see if the project has a valid directory
	err = util.IsValidProjectDir(path, location.DevfileLocation(""), fs)
	if err != nil {
		return err
	}

	if verbose {
		log.Info("\nStarter Project")
	}

	if starterProject.Git != nil {
		err := downloadGitProject(starterProject, decryptedToken, path, verbose)

		if err != nil {
			return err
		}

	} else if starterProject.Zip != nil {
		url := starterProject.Zip.Location
		sparseDir := starterProject.SubDir
		var downloadSpinner *log.Status
		if verbose {
			downloadSpinner = log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, url)
		}
		err := checkoutProject(sparseDir, url, path, decryptedToken, fs)
		if err != nil {
			if verbose {
				downloadSpinner.End(false)
			}
			return err
		}
		if verbose {
			downloadSpinner.End(true)
		}
	} else {
		return errors.New("project type not supported")
	}

	return nil
}

// downloadGitProject downloads the git starter projects from devfile.yaml
func downloadGitProject(starterProject *devfilev1.StarterProject, starterToken, path string, verbose bool) error {
	remoteName, remoteUrl, revision, err := parsercommon.GetDefaultSource(starterProject.Git.GitLikeProjectSource)
	if err != nil {
		return fmt.Errorf("unable to get default project source for starter project %s: %w", starterProject.Name, err)
	}

	// convert revision to referenceName type, ref name could be a branch or tag
	// if revision is not specified it would be the default branch of the project
	refName := plumbing.ReferenceName(revision)

	if plumbing.IsHash(revision) {
		// Specifying commit in the reference name is not supported by the go-git library
		// while doing git.PlainClone()
		log.Warning("Specifying commit in 'revision' is not yet supported in odo.")
		// overriding revision to empty as we do not support this
		revision = ""
	}

	if revision != "" {
		// lets consider revision to be a branch name first
		refName = plumbing.NewBranchReferenceName(revision)
	}

	var downloadSpinner *log.Status
	if verbose {
		downloadSpinner = log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, remoteUrl)
		defer downloadSpinner.End(false)
	}

	cloneOptions := &git.CloneOptions{
		URL:        remoteUrl,
		RemoteName: remoteName,
		// we don't need history for starter projects
		Depth: 1,
	}

	if refName != "" {
		cloneOptions.ReferenceName = refName
		cloneOptions.SingleBranch = true

	}

	if starterToken != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: RegistryUser,
			Password: starterToken,
		}
	}

	originalPath := ""
	if starterProject.SubDir != "" {
		originalPath = path
		path, err = os.MkdirTemp("", "")
		if err != nil {
			return err
		}
	}

	_, err = git.PlainClone(path, false, cloneOptions)

	if err != nil {

		// it returns the following error if no matching ref found
		// if we get this error, we are trying again considering revision as tag, only if revision is specified.
		if _, ok := err.(git.NoMatchingRefSpecError); !ok || revision == "" {
			return err
		}

		// try again to consider revision as tag name
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(revision)
		// remove if any .git folder downloaded in above try
		_ = os.RemoveAll(filepath.Join(path, ".git"))
		_, err = git.PlainClone(path, false, cloneOptions)
		if err != nil {
			return err
		}
	}

	// we don't want to download project be a git repo
	err = os.RemoveAll(filepath.Join(path, ".git"))
	if err != nil {
		// we don't need to return (fail) if this happens
		log.Warning("Unable to delete .git from cloned starter project")
	}

	if starterProject.SubDir != "" {
		err = util.GitSubDir(path, originalPath,
			starterProject.SubDir)
		if err != nil {
			return err
		}
	}
	if verbose {
		downloadSpinner.End(true)
	}

	return nil

}
