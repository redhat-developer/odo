//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// InteractiveContext represents the context of an interactive command to be run.
type InteractiveContext struct {

	//Command represents the original command ran
	Command []string

	// console is the internal interface used by the interactive command
	console *expect.Console

	// buffer is the internal bytes buffer containing the console output.
	// Its content will get updated as long as there are interactions with the console, like sending lines or
	// expecting lines.
	buffer *bytes.Buffer
}

// Tester represents the function that contains all steps to test the given interactive command.
// The InteractiveContext argument needs to be passed to the various helper.SendLine and helper.ExpectString methods.
type Tester func(InteractiveContext)

// RunInteractive runs the command in interactive mode and returns the output, and error.
// It takes command as array of strings, and a function `tester` that contains steps to run the test as an argument.
// The command is executed as a separate process, the environment of which is controlled via the `env` argument.
// The initial value of the sub-process environment is a copy of the environment of the current process.
// If `env` is not `nil`, it will be appended to the end of the sub-process environment.
// If there are duplicate environment keys, only the last value in the slice for each duplicate key is used.
func RunInteractive(command []string, env []string, tester Tester) (string, error) {

	fmt.Fprintln(GinkgoWriter, "running command", command, "with env", env)

	ptm, pts, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}

	term := vt10x.New(vt10x.WithWriter(pts))

	c, err := expect.NewConsole(expect.WithStdin(ptm), expect.WithStdout(term), expect.WithCloser(pts, ptm), expect.WithDefaultTimeout(3*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// execute the command
	cmd := exec.Command(command[0], command[1:]...)
	// setup stdin, stdout and stderr
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	ctx := InteractiveContext{
		Command: command,
		console: c,
		buffer:  buf,
	}
	tester(ctx)
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()

	return buf.String(), err
}

// expectDescriptionSupplier returns a function intended to be used as description supplier
// when checking errors do not occur in ExpectString and SendLine.
// Note that the function returned is evaluated lazily, only in case an error occurs.
func expectDescriptionSupplier(ctx InteractiveContext, line string) func() string {
	return func() string {
		return fmt.Sprintf("error while sending or expecting line: \"%s\"\n"+
			"=== output of command '%+q' read so far ===\n%v\n======================",
			line,
			ctx.Command,
			ctx.buffer)
	}
}

func SendLine(ctx InteractiveContext, line string) {
	_, err := ctx.console.SendLine(line)
	Expect(err).ShouldNot(HaveOccurred(), expectDescriptionSupplier(ctx, line))
}

func ExpectString(ctx InteractiveContext, line string) {
	res, err := ctx.console.ExpectString(line)
	fmt.Fprint(ctx.buffer, res)
	Expect(err).ShouldNot(HaveOccurred(), expectDescriptionSupplier(ctx, line))
}
