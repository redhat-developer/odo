package dev

import (
	"bytes"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"testing"
)

func TestDev_Start(t *testing.T) {
	d := Dev{}
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	devfileObj := parser.DevfileObj{
		Data: devfileData,
	}
	out := bytes.Buffer{}
	path := "/tmp"

	d.Start(devfileObj, &out, path)
}
