# go-expect

Package expect provides an expect-like interface to automate control of applications. It is unlike expect in that it does not spawn or manage process lifecycle. This package only focuses on expecting output and sending input through it's pseudoterminal.

This is a fork of the original repository [Netflix/go-expect](https://github.com/Netflix/go-expect) mostly to add Windows support. This fork has been added to test the [ActiveState state tool](https://www.activestate.com/products/platform/state-tool/)

Relevant additions:

- Windows support (Windows 10 and Windows Sever 2019 only)
- `expect.Console` is created with [xpty](https://github.com/ActiveState/termtest/xpty) allowing testing of applications that want to talk to an `xterm`-compatible terminal
- Filter out VT control characters in output. This is important for Windows support, as the windows pseudo-console creates lots of control-characters that can break up words.

See also [ActiveState/termtest](https://github.com/ActiveState/termtest) for a library that uses this package, but adds more life-cycle management.

## Usage

### `os.Exec` example

```go
package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/ActiveState/termtest/expect"
)

func main() {
	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	cmd := exec.Command("vi")
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	go func() {
		c.ExpectEOF()
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)
	c.Send("iHello world\x1b")
	time.Sleep(time.Second)
	c.Send("dd")
	time.Sleep(time.Second)
	c.SendLine(":q!")

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
```

### `golang.org/x/crypto/ssh/terminal` example

```
package main

import (
	"fmt"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/ActiveState/termtest/expect"
)

func getPassword(fd int) string {
	bytePassword, _ := terminal.ReadPassword(fd)

	return string(bytePassword)
}

func main() {
	c, _ := expect.NewConsole()

	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		c.SendLine("hunter2")
	}()

	echoText := getPassword(int(c.Tty().Fd()))

	<-donec

	fmt.Printf("\nPassword from stdin: %s", echoText)
}
```
