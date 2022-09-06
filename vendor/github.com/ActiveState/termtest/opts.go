// Copyright 2020 ActiveState Software. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file

package termtest

import (
	"io/ioutil"
	"os"
	"time"

	expect "github.com/ActiveState/termtest/expect"
)

// SendObserver is function that is called when text is send to the console
// Arguments are the message, number of bytes written and an error message
// See TestSendObserveFn for an example
type SendObserver func(msg string, num int, err error)

// Options contain optional values for ConsoleProcess construction and usage.
type Options struct {
	DefaultTimeout time.Duration
	WorkDirectory  string
	RetainWorkDir  bool
	Environment    []string
	ObserveSend    SendObserver
	ObserveExpect  expect.ExpectObserver
	CmdName        string
	Args           []string
	HideCmdLine    bool
	ExtraOpts      []expect.ConsoleOpt
}

// Normalize fills in default options
func (opts *Options) Normalize() error {
	if opts.DefaultTimeout == 0 {
		opts.DefaultTimeout = time.Second * 20
	}

	if opts.WorkDirectory == "" {
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			return err
		}
		opts.WorkDirectory = tmpDir
	}

	if opts.ObserveSend == nil {
		opts.ObserveSend = func(string, int, error) {}
	}

	if opts.ObserveExpect == nil {
		opts.ObserveExpect = func([]expect.Matcher, *expect.MatchState, error) {}
	}

	return nil
}

// CleanUp cleans up the environment
func (opts *Options) CleanUp() error {
	if !opts.RetainWorkDir {
		return os.RemoveAll(opts.WorkDirectory)
	}

	return nil
}
