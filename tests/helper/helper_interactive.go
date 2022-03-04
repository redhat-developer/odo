//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package helper

import (
	"bytes"
	"log"
	"os/exec"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
	. "github.com/onsi/gomega"
)

type Tester func(*expect.Console, *bytes.Buffer)

// RunInteractive runs the command in interactive mode and returns the output, and error.
// It takes command as array of strings, and a function `tester` that contains steps to run the test as an argument.
func RunInteractive(command []string, tester Tester) (string, error) {

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
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	tester(c, buf)
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()

	return buf.String(), err
}

func SendLine(c *expect.Console, line string) {
	_, err := c.SendLine(line)
	Expect(err).ShouldNot(HaveOccurred())
}

func ExpectString(c *expect.Console, line string) string {
	res, err := c.ExpectString(line)
	Expect(err).ShouldNot(HaveOccurred())
	return res
}
