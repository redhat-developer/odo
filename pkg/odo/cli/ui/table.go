package ui

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/redhat-developer/odo/pkg/log"
)

func NewTable() table.Writer {
	// Create the table and use our own style
	t := table.NewWriter()

	// Set the style of the table
	t.SetStyle(table.Style{
		Box: table.BoxStyle{
			PaddingLeft:  " ",
			PaddingRight: " ",
		},
		Color: table.ColorOptions{
			Header: text.Colors{text.FgHiGreen, text.Underline},
		},
		Format: table.FormatOptions{
			Footer: text.FormatUpper,
			Header: text.FormatUpper,
			Row:    text.FormatDefault,
		},
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: false,
			SeparateFooter:  false,
			SeparateHeader:  false,
			SeparateRows:    false,
		},
	})
	t.SetOutputMirror(log.GetStdout())
	return t
}
