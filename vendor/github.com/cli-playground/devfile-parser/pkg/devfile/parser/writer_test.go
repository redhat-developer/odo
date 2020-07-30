package parser

import (
	"testing"

	devfileCtx "github.com/cli-playground/devfile-parser/pkg/devfile/parser/context"
	v100 "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/1.0.0"
	"github.com/cli-playground/devfile-parser/pkg/testingutil/filesystem"
)

func TestWriteJsonDevfile(t *testing.T) {

	var (
		devfileTempPath = "devfile.yaml"
		apiVersion      = "1.0.0"
		testName        = "TestName"
	)

	t.Run("write json devfile", func(t *testing.T) {

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
			Data: &v100.Devfile100{
				ApiVersion: v100.ApiVersion(apiVersion),
				Metadata: v100.Metadata{
					Name: testName,
				},
			},
		}

		// Use fakeFs
		fs := filesystem.NewFakeFs()
		devfileObj.Ctx.Fs = fs

		// test func()
		err := devfileObj.WriteJsonDevfile()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}

		if _, err := fs.Stat(OutputDevfileJsonPath); err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})

	t.Run("write yaml devfile", func(t *testing.T) {

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
			Data: &v100.Devfile100{
				ApiVersion: v100.ApiVersion(apiVersion),
				Metadata: v100.Metadata{
					Name: testName,
				},
			},
		}

		// Use fakeFs
		fs := filesystem.NewFakeFs()
		devfileObj.Ctx.Fs = fs

		// test func()
		err := devfileObj.WriteYamlDevfile()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}

		if _, err := fs.Stat(OutputDevfileYamlPath); err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})
}
