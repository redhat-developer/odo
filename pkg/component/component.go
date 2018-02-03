package component

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

const applicationLabel = "app.ocdev.developers.redhat.com"
const componentLabel = "component.ocdev.developers.redhat.com"

func CreateFromGit(name string, ctype string, url string) (string, error) {
	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current application")
	}

	output, err := occlient.NewAppS2I(name, ctype, url, map[string]string{applicationLabel: currentAppliction, componentLabel: name})
	if err != nil {
		return "", err
	}

	return output, nil
}

func CreateEmpty(name string, ctype string) (string, error) {
	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current application")
	}

	output, err := occlient.NewAppS2IEmpty(name, ctype, map[string]string{applicationLabel: currentAppliction, componentLabel: name})
	if err != nil {
		return "", err
	}

	return output, nil
}

func CreateFromDir(name string, ctype, dir string) (string, error) {
	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current application")
	}

	output, err := occlient.NewAppS2IEmpty(name, ctype, map[string]string{applicationLabel: currentAppliction, componentLabel: name})
	if err != nil {
		return "", err
	}

	// TODO: it might not be ideal to print to stdout here
	fmt.Println(output)
	fmt.Println("please wait, building application...")

	output, err = occlient.StartBuildFromDir(name, dir)
	if err != nil {
		return "", err
	}
	fmt.Println(output)

	return "", nil

}
