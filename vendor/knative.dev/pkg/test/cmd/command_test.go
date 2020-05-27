/*
Copyright 2019 The Knative Authors

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

package cmd

import (
	"reflect"
	"testing"
)

func TestRunCommand(t *testing.T) {
	testCases := []struct {
		command             string
		options             []Option
		expectedOutput      string
		expectedErrorOutput string
		expectedErrorCode   int
	}{{
		"",
		[]Option{},
		"",
		invalidInputErrorPrefix + "",
		1,
	}, {
		" ",
		[]Option{},
		"",
		invalidInputErrorPrefix + " ",
		1,
	}, {
		"echo hello, world",
		[]Option{},
		"hello, world\n",
		"",
		0,
	}, {
		"bash -c 'echo foo > /dev/stderr; exit 4'",
		[]Option{},
		"",
		"foo\n",
		4,
	}, {
		"bash -c 'echo ${HELLO} > /dev/stdout; exit 0'",
		[]Option{WithEnvs([]string{"HELLO=hello, world"})},
		"hello, world\n",
		"",
		0,
	}}
	for _, c := range testCases {
		out, err := RunCommand(c.command, c.options...)
		if c.expectedOutput != out {
			t.Fatalf("Expect %q but actual is %q", c.expectedOutput, out)
		}
		if err != nil {
			if ce, ok := err.(*CommandLineError); ok {
				if ce.ErrorCode != c.expectedErrorCode {
					t.Fatalf("Expect to get error code %d but got %d", c.expectedErrorCode, ce.ErrorCode)
				}
				if string(ce.ErrorOutput) != c.expectedErrorOutput {
					t.Fatalf("Expect to get error message %q but got %q", c.expectedErrorOutput, ce.ErrorOutput)
				}
			} else {
				t.Fatalf("Expect to get a CommandLineError but got %q", reflect.TypeOf(err))
			}
		} else {
			if c.expectedErrorCode != 0 {
				t.Fatalf("Expect to get an error code %d but got no error", c.expectedErrorCode)
			}
			if c.expectedErrorOutput != "" {
				t.Fatalf("Expect to get error message %q but got nothing", c.expectedErrorOutput)
			}
		}
	}
}

func TestRunCommands(t *testing.T) {
	testCases := []struct {
		commands            []string
		expectedOutput      string
		expectedErrorOutput string
		expectedErrorCode   int
	}{
		{
			[]string{"echo 123", "echo 234", "echo 345"},
			"123\n\n234\n\n345\n",
			"",
			0,
		},
		{
			[]string{" ", "echo 123"},
			"",
			invalidInputErrorPrefix + " ",
			1,
		},
		{
			[]string{"echo 123", "", "echo 234"},
			"123\n\n",
			invalidInputErrorPrefix + "",
			1,
		},
		{
			[]string{`bash -c "echo foo > /dev/stderr; exit 4"`},
			"",
			"foo\n",
			4,
		},
		{
			[]string{"bash -c 'exit 10'", "echo 123"},
			"",
			"",
			10,
		},
	}
	for _, c := range testCases {
		out, err := RunCommands(c.commands...)
		if c.expectedOutput != out {
			t.Fatalf("Expect %q but actual is %q", c.expectedOutput, out)
		}
		if err != nil {
			if ce, ok := err.(*CommandLineError); ok {
				if ce.ErrorCode != c.expectedErrorCode {
					t.Fatalf("Expect to get error code %d but got %d", c.expectedErrorCode, ce.ErrorCode)
				}
				if string(ce.ErrorOutput) != c.expectedErrorOutput {
					t.Fatalf("Expect to get error message %q but got %q", c.expectedErrorOutput, ce.ErrorOutput)
				}
			} else {
				t.Fatalf("Expect to get a CommandLineError but got %s", reflect.TypeOf(err))
			}
		} else {
			if c.expectedErrorCode != 0 {
				t.Fatalf("Expect to get an error code %d but got no error", c.expectedErrorCode)
			}
			if c.expectedErrorOutput != "" {
				t.Fatalf("Expect to get error message %q but got nothing", c.expectedErrorOutput)
			}
		}
	}
}

func TestRunCommandsInParallel(t *testing.T) {
	testCases := []struct {
		commands            []string
		possibleOutput      []string
		possibleErrorOutput []string
	}{
		{
			[]string{"echo 123", "echo 234"},
			[]string{"123\n\n234\n", "234\n\n123\n"},
			nil,
		},
		{
			[]string{"", "echo 123"},
			[]string{"\n123\n", "123\n\n"},
			[]string{invalidInputErrorPrefix + "", invalidInputErrorPrefix + ""},
		},
		{
			[]string{"bash -c 'echo foo; exit 1'", "bash -c 'echo bar > /dev/stderr; exit 1'"},
			[]string{"\nfoo\n", "foo\n\n"},
			[]string{"bar\n\n", "\nbar\n"},
		},
	}
	for _, c := range testCases {
		out, err := RunCommandsInParallel(c.commands...)

		idx := -1
		for i := range c.possibleOutput {
			if c.possibleOutput[i] == out {
				idx = i
				break
			}
		}
		if idx == -1 {
			t.Fatalf("Expect output in %v but actual is %q", c.possibleOutput, out)
		}

		if len(c.possibleErrorOutput) != 0 {
			if err != nil {
				if err.Error() != c.possibleErrorOutput[idx] {
					t.Fatalf("Got an error %q but should get %q", err.Error(), c.possibleErrorOutput[idx])
				}
			} else {
				t.Fatalf("Expect to get an error %q but got nil", c.possibleErrorOutput[idx])
			}
		} else {
			if err != nil {
				t.Fatalf("Expect to get no error but got %v", err)
			}
		}
	}
}
