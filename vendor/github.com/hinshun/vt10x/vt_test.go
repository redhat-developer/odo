package vt10x

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh/terminal"
)

func extractStr(t *State, x0, x1, row int) string {
	var s []rune
	for i := x0; i <= x1; i++ {
		c, _, _ := t.Cell(i, row)
		s = append(s, c)
	}
	return string(s)
}

func TestPlainChars(t *testing.T) {
	var st State
	term, err := Create(&st, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello world!"
	_, err = term.Write([]byte(expected))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	actual := extractStr(&st, 0, len(expected)-1, 0)
	if expected != actual {
		t.Fatal(actual)
	}
}

func TestNewline(t *testing.T) {
	var st State
	term, err := Create(&st, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello world!\n...and more."
	_, err = term.Write([]byte("\033[20h")) // set CRLF mode
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	_, err = term.Write([]byte(expected))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	split := strings.Split(expected, "\n")
	actual := extractStr(&st, 0, len(split[0])-1, 0)
	actual += "\n"
	actual += extractStr(&st, 0, len(split[1])-1, 1)
	if expected != actual {
		t.Fatal(actual)
	}

	// A newline with a color set should not make the next line that color,
	// which used to happen if it caused a scroll event.
	st.moveTo(0, st.rows-1)
	_, err = term.Write([]byte("\033[1;37m\n$ \033[m"))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	_, fg, bg := st.Cell(st.Cursor())
	if fg != DefaultFG {
		t.Fatal(st.cur.x, st.cur.y, fg, bg)
	}
}

var (
	dsrPattern = regexp.MustCompile(`(\d+);(\d+)`)
)

type Coord struct {
	row int
	col int
}

func TestVTCPR(t *testing.T) {
	c, _, err := NewVT10XConsole()
	require.NoError(t, err)
	defer c.Close()

	go func() {
		c.ExpectEOF()
	}()

	coord, err := cpr(c.Tty())
	require.NoError(t, err)
	require.Equal(t, 1, coord.row)
	require.Equal(t, 1, coord.col)
}

// cpr is an example application that requests for the cursor position report.
func cpr(tty *os.File) (*Coord, error) {
	oldState, err := terminal.MakeRaw(int(tty.Fd()))
	if err != nil {
		return nil, err
	}

	defer terminal.Restore(int(tty.Fd()), oldState)

	// ANSI escape sequence for DSR - Device Status Report
	// https://en.wikipedia.org/wiki/ANSI_escape_code#CSI_sequences
	fmt.Fprint(tty, "\x1b[6n")

	// Reports the cursor position (CPR) to the application as (as though typed at
	// the keyboard) ESC[n;mR, where n is the row and m is the column.
	reader := bufio.NewReader(tty)
	text, err := reader.ReadSlice('R')
	if err != nil {
		return nil, err
	}

	matches := dsrPattern.FindStringSubmatch(string(text))
	if len(matches) != 3 {
		return nil, fmt.Errorf("incorrect number of matches: %d", len(matches))
	}

	col, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	row, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	return &Coord{row, col}, nil
}
