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

package interactive

import (
	"os/exec"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

var (
	mySpew        *spew.ConfigState
	arbitraryArgs []string
)

func init() {
	mySpew = spew.NewDefaultConfig()
	mySpew.DisableMethods = true
	arbitraryArgs = []string{"docker", "run", "something"}
}

func argsValidator(t *testing.T, args []string) func(c *exec.Cmd) error {
	return func(c *exec.Cmd) error {
		for i, arg := range arbitraryArgs {
			if c.Args[i] != arg {
				t.Errorf("c.Args[%d] not correct", i)
			}
		}
		t.Log("Validated c.Args")
		if t.Failed() {
			mySpew.Sdump(*c)
		}
		return nil
	}
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(arbitraryArgs...)
	cmd.run = argsValidator(t, arbitraryArgs)
	cmd.Run()
}

func TestAddArgs(t *testing.T) {
	cmd := NewCommand(arbitraryArgs[0])
	cmd.AddArgs(arbitraryArgs[1:]...)
	cmd.run = argsValidator(t, arbitraryArgs)
	cmd.Run()
}

func TestCommandStringer(t *testing.T) {
	t.Log(NewCommand("true"))
}
