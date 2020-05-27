/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Helper functions for running interactive CLI sessions from Go
package interactive

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// could be public so extensions of this can test themselves
var defaultRun func(*exec.Cmd) error

func init() {
	defaultRun = func(c *exec.Cmd) error { return c.Run() }
}

// Command represents running an executable, to make it easy to write interactive "shell scripts" in Go.
// This routes standard in, out, and error from and to the calling program, intended usually to be a login shell run by a person.
type Command struct {
	Name string
	Args []string
	// If LogFile is set, anything send to stdout and stderr is tee'd to the file named in LogFile when .Run() is called.
	LogFile string
	run     func(*exec.Cmd) error
}

// NewCommand creates an Command, breaking out .Name and .Args for you
func NewCommand(cmdAndArgs ...string) Command {
	if len(cmdAndArgs) < 1 {
		log.Fatal("NewCommand must have a least the command given")
	}
	return Command{cmdAndArgs[0], cmdAndArgs[1:], "", defaultRun}
}

// Run executes the command with exec.Command(), sets standard in/out/err, and returns the result of exec.Cmd.Run()
// Will tee stdout&err to .LogFile if set
func (c Command) Run() error {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = os.Stdin
	if len(c.LogFile) > 0 {
		f, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		cmd.Stdout = io.MultiWriter(os.Stdout, f)
		cmd.Stderr = io.MultiWriter(os.Stderr, f)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return c.run(cmd)
}

// String conforms to Stringer
func (c Command) String() string {
	s := c.Name + " " + strings.Join(c.Args, " ")
	return "Command to run: " + s
}

// AddArgs appends arguments to the current list of args
func (c *Command) AddArgs(args ...string) {
	c.Args = append(c.Args, args...)
}
