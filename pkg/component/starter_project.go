package component

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	"github.com/openshift/odo/pkg/util"

	"github.com/pkg/errors"
)

const (
	defaultStarterProjectName = "devfile-starter-project-name"
)

func checkoutProject(subDir, zipURL, path, starterToken string) error {

	if subDir == "" {
		subDir = "/"
	}
	err := util.GetAndExtractZip(zipURL, path, subDir, starterToken)
	if err != nil {
		return errors.Wrap(err, "failed to download and extract project zip folder")
	}
	return nil
}

// GetStarterProject gets starter project value from flag --starter.
func GetStarterProject(projects []devfilev1.StarterProject, projectPassed string) (project *devfilev1.StarterProject, err error) {

	nOfProjects := len(projects)

	if nOfProjects == 0 {
		return nil, errors.Errorf("no starter project found in devfile.")
	}

	// Determine what project to be used
	if nOfProjects == 1 && projectPassed == defaultStarterProjectName {
		project = &projects[0]
	} else if nOfProjects > 1 && projectPassed == defaultStarterProjectName {
		project = &projects[0]
		log.Warning("There are multiple projects in this devfile but none have been specified in --starter. Downloading the first: " + project.Name)
	} else { //If the user has specified a project
		var availableNames []string

		projectFound := false
		for indexOfProject, projectInfo := range projects {
			availableNames = append(availableNames, projectInfo.Name)
			if projectInfo.Name == projectPassed { //Get the index
				project = &projects[indexOfProject]
				projectFound = true
			}
		}

		if !projectFound {
			availableNamesString := strings.Join(availableNames, ",")
			return nil, errors.Errorf("the project: %s specified in --starter does not exist, available projects: %s", projectPassed, availableNamesString)
		}
	}

	return project, err

}

// DownloadStarterProject Downloads first starter project from list of starter projects in devfile
func DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string) error {
	var path string
	var err error
	// Retrieve the working directory in order to clone correctly
	if contextDir == "" {
		path, err = os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "Could not get the current working directory.")
		}
	} else {
		path = contextDir
	}

	// We will check to see if the project has a valid directory
	err = util.IsValidProjectDir(path, location.DevfileLocation(""))
	if err != nil {
		return err
	}

	log.Info("\nStarter Project")

	if starterProject.Git != nil {
		err := downloadGitProject(starterProject, decryptedToken, path)

		if err != nil {
			return err
		}

	} else if starterProject.Zip != nil {
		url := starterProject.Zip.Location
		sparseDir := starterProject.SubDir
		downloadSpinner := log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, url)
		err := checkoutProject(sparseDir, url, path, decryptedToken)
		if err != nil {
			downloadSpinner.End(false)
			return err
		}
		downloadSpinner.End(true)
	} else {
		return errors.Errorf("Project type not supported")
	}

	return nil
}

// downloadGitProject downloads the git starter projects from devfile.yaml
func downloadGitProject(starterProject *devfilev1.StarterProject, starterToken, path string) error {
	remoteName, remoteUrl, revision, err := parsercommon.GetDefaultSource(starterProject.Git.GitLikeProjectSource)
	if err != nil {
		return errors.Wrapf(err, "unable to get default project source for starter project %s", starterProject.Name)
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

	downloadSpinner := log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, remoteUrl)
	defer downloadSpinner.End(false)

	cloneOptions := &git.CloneOptions{
		URL:           remoteUrl,
		RemoteName:    remoteName,
		ReferenceName: refName,
		SingleBranch:  true,
		// we don't need history for starter projects
		Depth: 1,
	}

	if starterToken != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: registryUtil.RegistryUser,
			Password: starterToken,
		}
	}

	originalPath := ""
	if starterProject.SubDir != "" {
		originalPath = path
		path, err = ioutil.TempDir("", "")
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
	downloadSpinner.End(true)

	return nil

}
