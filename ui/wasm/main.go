package main

import (
	"syscall/js"

	"github.com/feloy/devfile-builder/wasm/pkg/exports"
)

func setFreshDevfile() {
	content := `schemaVersion: 2.2.0`
	exports.SetDevfileContent(content)
}

func main() {
	setFreshDevfile()

	js.Global().Set("addContainer", js.FuncOf(exports.AddContainerWrapper))
	js.Global().Set("addImage", js.FuncOf(exports.AddImageWrapper))
	js.Global().Set("addResource", js.FuncOf(exports.AddResourceWrapper))

	js.Global().Set("addExecCommand", js.FuncOf(exports.AddExecCommandWrapper))
	js.Global().Set("addApplyCommand", js.FuncOf(exports.AddApplyCommandWrapper))
	js.Global().Set("addCompositeCommand", js.FuncOf(exports.AddCompositeCommandWrapper))

	js.Global().Set("getFlowChart", js.FuncOf(exports.GetFlowChartWrapper))
	js.Global().Set("setDevfileContent", js.FuncOf(exports.SetDevfileContentWrapper))

	js.Global().Set("setMetadata", js.FuncOf(exports.SetMetadataWrapper))

	js.Global().Set("moveCommand", js.FuncOf(exports.MoveCommandWrapper))

	js.Global().Set("setDefaultCommand", js.FuncOf(exports.SetDefaultCommandWrapper))
	js.Global().Set("unsetDefaultCommand", js.FuncOf(exports.UnsetDefaultCommandWrapper))

	js.Global().Set("deleteCommand", js.FuncOf(exports.DeleteCommandWrapper))
	js.Global().Set("deleteContainer", js.FuncOf(exports.DeleteContainerWrapper))
	js.Global().Set("deleteImage", js.FuncOf(exports.DeleteImageWrapper))
	js.Global().Set("deleteResource", js.FuncOf(exports.DeleteResourceWrapper))

	js.Global().Set("updateEvents", js.FuncOf(exports.UpdateEventsWrapper))

	js.Global().Set("isQuantityValid", js.FuncOf(exports.IsQuantityValidWrapper))
	<-make(chan bool)
}
