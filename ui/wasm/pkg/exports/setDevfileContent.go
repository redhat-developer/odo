package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"

	"k8s.io/utils/pointer"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

// setDevfileContent

func SetDevfileContentWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		SetDevfileContent(args[0].String()),
	)
}

func SetDevfileContent(content string) (map[string]interface{}, error) {
	parserArgs := parser.ParserArgs{
		Data:                          []byte(content),
		ConvertKubernetesContentInUri: pointer.Bool(false),
	}
	var err error
	global.Devfile, _, err = devfile.ParseDevfileAndValidate(parserArgs)
	if err != nil {
		return nil, fmt.Errorf("error parsing devfile YAML: %w", err)
	}
	global.FS = filesystem.NewFakeFs()
	global.Devfile.Ctx = context.FakeContext(global.FS, "/devfile.yaml")

	return utils.GetContent()
}
