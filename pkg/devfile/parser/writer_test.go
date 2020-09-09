package parser

import (
	"testing"

	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestWriteJsonDevfile(t *testing.T) {

	var (
		schemaVersion = "2.0.0"
		testName      = "TestName"
	)

	t.Run("write json devfile", func(t *testing.T) {

		// Use fakeFs
		fs := filesystem.NewFakeFs()

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.FakeContext(fs, OutputDevfileJsonPath),
			Data: &v200.Devfile200{
				SchemaVersion: schemaVersion,
				Metadata: common.DevfileMetadata{
					Name: testName,
				},
			},
		}

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

		// Use fakeFs
		fs := filesystem.NewFakeFs()

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
			Data: &v200.Devfile200{
				SchemaVersion: schemaVersion,
				Metadata: common.DevfileMetadata{
					Name: testName,
				},
			},
		}

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
