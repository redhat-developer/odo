package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"

	"github.com/redhat-developer/odo/pkg/api"
)

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
