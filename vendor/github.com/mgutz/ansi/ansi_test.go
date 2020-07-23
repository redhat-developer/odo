package ansi

import (
	"fmt"
	"strings"
	"testing"
)

func TestPlain(t *testing.T) {
	DisableColors(true)
	PrintStyles()
}

func TestStyles(t *testing.T) {
	DisableColors(false)
	PrintStyles()
}

func TestDisableColors(t *testing.T) {
	fn := ColorFunc("red")

	buf := colorCode("off")
	if buf.String() != "" {
		t.Fail()
	}

	DisableColors(true)
	if Black != "" {
		t.Fail()
	}
	code := ColorCode("red")
	if code != "" {
		t.Fail()
	}
	s := fn("foo")
	if s != "foo" {
		t.Fail()
	}

	DisableColors(false)
	if Black == "" {
		t.Fail()
	}
	code = ColorCode("red")
	if code == "" {
		t.Fail()
	}
	// will have escape codes around it
	index := strings.Index(fn("foo"), "foo")
	if index <= 0 {
		t.Fail()
	}
}

func TestAttributeReset(t *testing.T) {
	boldRed := ColorCode("red+b")
	greenUnderline := ColorCode("green+u")
	s := fmt.Sprintf("normal %s bold red %s green underline %s", boldRed, greenUnderline, Reset)
	// See the results on the terminal for regression tests.
	fmt.Printf("Colored string: %s\n", s)
	fmt.Printf("Escaped string: %q\n", s)
	if s != "normal \x1b[0;1;31m bold red \x1b[0;4;32m green underline \x1b[0m" {
		t.Error("Attributes are not being reset")
	}
}
