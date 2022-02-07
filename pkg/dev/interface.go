package dev

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"io"
)

type Client interface {
	//GetComponents() (devfile.Component, error)
	Start(parser.DevfileObj, io.Writer, string) error
	Cleanup() error
}
