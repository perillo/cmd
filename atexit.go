// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"sync"
)

var atExitFuncs []func()

// AtExit will call f when Exit is called.
func AtExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

// Exit calls os.Exit with the exit status as set by SetExitStatus.  It calls
// all the function registered by AtExit in FIFO order.
func Exit() {
	for _, f := range atExitFuncs {
		f()
	}

	os.Exit(exitStatus)
}

// Fatalf prints the formatted message on os.Stderr and exit with exit status
// 1.
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	Exit()
}

// Errorf prints the formatted message on os.Stderr and set the exit status to
// 1.
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	SetExitStatus(1)
}

// ExitIfErrors will exit if the current exit status is not 0.
func ExitIfErrors() {
	if exitStatus != 0 {
		Exit()
	}
}

var exitMu sync.Mutex // guards exitStatus
var exitStatus = 0

// SetExitStatus sets the exit status to n.
func SetExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

// GetExitStatus returns the current exit status.
func GetExitStatus() int {
	return exitStatus
}
