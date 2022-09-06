// Copyright 2020 ActiveState Software. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file

// +build linux darwin

package osutils

import "syscall"

// SysProcAttrForNewProcessGroup returns a SysProcAttr structure configured to start a process with a new process group
func SysProcAttrForNewProcessGroup() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
