package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	. "github.com/onsi/ginkgo/v2"

	"github.com/redhat-developer/odo/pkg/api"
)

// Version pattern has always been in the form of X.X.X
var versionRe = regexp.MustCompile(`(\d.\d.\d)`)

type Registry struct {
	url string
}

func NewRegistry(url string) Registry {
	return Registry{
		url: url,
	}
}

func (o Registry) GetIndex() ([]api.DevfileStack, error) {
	url, err := url.JoinPath(o.url, "v2index")
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			fmt.Fprintf(GinkgoWriter, "[warn] error closing response body: %v\n", cErr)
		}
	}()

	var target []api.DevfileStack
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (o Registry) GetStack(name string) (api.DevfileStack, error) {
	index, err := o.GetIndex()
	if err != nil {
		return api.DevfileStack{}, err
	}
	for _, stack := range index {
		if stack.Name == name {
			return stack, nil
		}
	}
	return api.DevfileStack{}, fmt.Errorf("stack %q not found", name)
}

// GetVersions returns the list of all versions for the given stack name in the given Devfile registry.
// It uses the "odo registry" command to find out this information.
//
// The registry name is optional, and defaults to DefaultDevfileRegistry if not set.
func GetVersions(registryName string, stackName string) []string {
	devfileReg := "DefaultDevfileRegistry"
	if registryName != "" {
		devfileReg = registryName
	}
	out := Cmd("odo", "registry", "--devfile", stackName, "--devfile-registry", devfileReg).ShouldPass().Out()
	return versionRe.FindAllString(out, -1)
}

// HasAtLeastTwoVersions returns whether the given stack in the given Devfile registry has at least two versions.
// This is useful to determine if the "Select version" prompt will be displayed in the interactive "odo init" tests.
// Otherwise, "odo init" would just skip this "Select version" prompt if the stack selected has no version or only a single one.
//
// Note that the registry name is optional, and defaults to DefaultDevfileRegistry if not set.
func HasAtLeastTwoVersions(registryName string, stackName string) bool {
	return len(GetVersions(registryName, stackName)) >= 2
}
