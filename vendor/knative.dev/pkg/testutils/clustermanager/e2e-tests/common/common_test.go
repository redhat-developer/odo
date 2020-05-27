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

package common

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestStandardExec(t *testing.T) {
	datas := []struct {
		cmd    string
		args   []string
		expOut []byte
		expErr error
	}{
		{"bash", []string{"-c", "echo foo"}, []byte("foo\n"), nil},
		{"cmd_not_exist", []string{"-c", "echo"}, []byte{}, errors.New(`exec: "cmd_not_exist": executable file not found in $PATH`)},
	}

	for _, data := range datas {
		t.Run(data.cmd, func(t *testing.T) {
			out, err := StandardExec(data.cmd, data.args...)
			errMsg := fmt.Sprintf("running cmd: '%v', args: '%v'", data.cmd, data.args)
			if (err != nil) != (data.expErr != nil) {
				t.Fatalf("Error = %v, want: %v", err, data.expErr)
			} else if err != nil {
				if err.Error() != data.expErr.Error() {
					t.Errorf("%s\nerror got:  '%v'\nerror want: '%v'", errMsg, data.expErr, err)
				}
			}

			if got, want := string(out), string(data.expOut); got != want {
				t.Errorf("%s\noutput got:  '%s'\noutput want: '%s'", errMsg, got, want)
			}
		})
	}
}

func TestGetRepoName(t *testing.T) {
	datas := []struct {
		out    string
		err    error
		expOut string
		expErr error
	}{
		{
			// Good run
			"a/b/c", nil, "c", nil,
		}, {
			// Good run
			"a/b/c/", nil, "c", nil,
		}, {
			// Git error
			"", fmt.Errorf("git error"), "", fmt.Errorf("failed git rev-parse --show-toplevel: 'git error'"),
		},
	}

	oldFunc := StandardExec
	defer func() {
		// restore
		StandardExec = oldFunc
	}()

	for _, data := range datas {
		// mock for testing
		StandardExec = func(name string, args ...string) ([]byte, error) {
			return []byte(data.out), data.err
		}

		out, err := GetRepoName()
		if data.expOut != out || !reflect.DeepEqual(err, data.expErr) {
			t.Errorf("testing getting repo name with:\n\tmocked git output: '%s'\n\tmocked git err: '%v'\nwant: out - '%s', err - '%v'\ngot: out - '%s', err - '%v'",
				data.out, data.err, data.expOut, data.expErr, out, err)
		}
	}
}
