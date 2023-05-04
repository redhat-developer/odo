package helper

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	resp, err := http.Get(o.url + "/v2index")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
