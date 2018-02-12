package component

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/config"
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

	if err = SetCurrent(name); err != nil {
		return "", errors.Wrapf(err, "unable to activate component %s created from git", name)
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
	if err = SetCurrent(name); err != nil {
		return "", errors.Wrapf(err, "unable to activate empty component %s", name)
	}

	return output, nil
}

func CreateFromDir(name string, ctype, dir string) (string, error) {
	output, err := CreateEmpty(name, ctype)
	if err != nil {
		return "", errors.Wrap(err, "unable to get create empty component")
	}

	// TODO: it might not be ideal to print to stdout here
	fmt.Println(output)
	fmt.Println("please wait, building application...")

	output, err = occlient.StartBuild(name, dir)
	if err != nil {
		return "", err
	}
	fmt.Println(output)

	return "", nil

}

// Delete whole component
func Delete(name string) (string, error) {
	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	output, err := occlient.Delete("all", "", map[string]string{applicationLabel: currentAppliction, componentLabel: name})
	if err != nil {
		return "", errors.Wrap(err, "unable to delete component")
	}

	return output, nil
}

func SetCurrent(name string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "unable to get config")
	}

	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return errors.Wrap(err, "unable to get current application")
	}

	cfg.SetActiveComponent(name, currentAppliction)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	return nil
}

func GetCurrent() (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentAppliction, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	currentComponent := cfg.GetActiveComponent(currentAppliction)
	if currentComponent == "" {
		return "", errors.New("no component is set as active")
	}
	return currentComponent, nil

}

func Push(name string, dir string) (string, error) {
	output, err := occlient.StartBuild(name, dir)
	if err != nil {
		return "", errors.Wrap(err, "unable to start build")
	}
	return output, nil
}
