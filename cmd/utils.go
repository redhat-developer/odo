package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
)

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string, currentProject string) error {
	componentList, err := component.List(client, appName, currentProject)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	for _, cmpnt := range componentList {
		_, _, componentURL, appStore, err := component.GetComponentDesc(client, cmpnt.Name, appName, currentProject)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		fmt.Println("Component", cmpnt.Name, "will be deleted.")

		if len(componentURL) != 0 {
			fmt.Println("  This component is externally exposed, and the URL will be removed")
		}

		for _, store := range appStore {
			fmt.Println("  This Component uses storage", store.Name, "of size", store.Size, "will be removed")
		}

	}
	return nil
}
