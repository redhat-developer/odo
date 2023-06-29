package helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ActiveState/termtest"
	"github.com/ActiveState/termtest/expect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// InteractiveContext represents the context of an interactive command to be run.
type InteractiveContext struct {

	// Command represents the original command ran
	Command []string

	// cp is the internal interface used by the interactive command
	cp *termtest.ConsoleProcess

	// buffer is the internal bytes buffer containing the console output.
	// Its content will get updated as long as there are interactions with the console, like sending lines or
	// expecting lines.
	buffer *bytes.Buffer

	// A function to call to stop the process
	StopCommand func()
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

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	opts := termtest.Options{
		CmdName:       command[0],
		Args:          command[1:],
		WorkDirectory: wd,
		RetainWorkDir: true,
		ExtraOpts:     []expect.ConsoleOpt{},
	}

	if env != nil {
		opts.Environment = append(os.Environ(), env...)
	}

	cp, err := termtest.New(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cp.Close()

	buf := new(bytes.Buffer)
	ctx := InteractiveContext{
		Command: command,
		buffer:  buf,
		StopCommand: func() {
			_ = cp.Signal(os.Kill)
		},
		cp: cp,
	}
	tester(ctx)

	_, err = cp.ExpectExitCode(0)

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
	ctx.cp.Send(line)
}

func PressKey(ctx InteractiveContext, c byte) {
	ctx.cp.SendUnterminated(string(c))
}

func ExpectString(ctx InteractiveContext, line string) {
	res, err := ctx.cp.Expect(line, 120*time.Second)
	fmt.Fprint(ctx.buffer, res)
	Expect(err).ShouldNot(HaveOccurred(), expectDescriptionSupplier(ctx, line))
}
