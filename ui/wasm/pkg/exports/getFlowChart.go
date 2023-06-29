package exports

import (
	"errors"
	"syscall/js"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-lifecycle/pkg/graph"
)

// getFlowChart

func GetFlowChartWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		getFlowChart(),
	)
}

func getFlowChart() (string, error) {
	g, err := graph.Build(global.Devfile.Data)
	if err != nil {
		return "", errors.New("error building graph")
	}
	return g.ToFlowchart().String(), nil
}
