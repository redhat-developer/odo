// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package helper

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
	. "github.com/onsi/gomega"
)

func RunInteractive(commonVar CommonVar, command []string, test func(*expect.Console, *bytes.Buffer)) (string, error) {

	ptm, pts, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}

	term := vt10x.New(vt10x.WithWriter(pts))

	c, err := expect.NewConsole(expect.WithStdin(ptm), expect.WithStdout(term), expect.WithCloser(pts, ptm))
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
	test(c, buf)
	if err != nil {
		log.Fatal(err)
	}
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
