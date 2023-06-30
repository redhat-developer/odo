package devstate

import (
	"errors"

	"github.com/feloy/devfile-lifecycle/pkg/graph"
)

func (o *DevfileState) GetFlowChart() (string, error) {
	g, err := graph.Build(o.Devfile.Data)
	if err != nil {
		return "", errors.New("error building graph")
	}
	return g.ToFlowchart().String(), nil
}
