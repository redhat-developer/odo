package testingutil

import (
	"bytes"
	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"testing"
)

// This whole file copies the testing infrastructure from survey lib since it cannot be imported. This mixes elements from:
// vendor/gopkg.in/AlecAivazis/survey.v1/survey_posix_test.go
// vendor/gopkg.in/AlecAivazis/survey.v1/survey_test.go
// vendor/gopkg.in/AlecAivazis/survey.v1/survey.go

type wantsStdio interface {
	WithStdio(terminal.Stdio)
}

func Stdio(c *expect.Console) terminal.Stdio {
	return terminal.Stdio{c.Tty(), c.Tty(), c.Tty()}
}

type PromptTest struct {
	name      string
	prompt    survey.Prompt
	procedure func(*expect.Console)
	expected  interface{}
}

func RunPromptTest(t *testing.T, test PromptTest) {
	var answer interface{}
	RunTest(t, test.procedure, func(stdio terminal.Stdio) error {
		var err error
		if p, ok := test.prompt.(wantsStdio); ok {
			p.WithStdio(stdio)
		}
		answer, err = test.prompt.Prompt()
		return err
	})
	require.Equal(t, test.expected, answer)
}

func RunTest(t *testing.T, procedure func(*expect.Console), test func(terminal.Stdio) error) {
	t.Parallel()

	// Multiplex output to a buffer as well for the raw bytes.
	buf := new(bytes.Buffer)
	c, state, err := vt10x.NewVT10XConsole(expect.WithStdout(buf))
	require.Nil(t, err)
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(c)
	}()

	err = test(Stdio(c))
	require.Nil(t, err)

	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()
	<-donec

	t.Logf("Raw output: %q", buf.String())

	// Dump the terminal's screen.
	t.Logf("\n%s", expect.StripTrailingEmptyLines(state.String()))
}
