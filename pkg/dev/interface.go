package dev

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"io"
)

type Client interface {
	Start(parser.DevfileObj, io.Writer, string) error
	Cleanup() error
}
