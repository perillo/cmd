// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmdstate

import "testing"

// TestErrorf tests that a call to Errorf sets the exit status to 1.
func TestErrorf(t *testing.T) {
	Errorf("error")
	if code := GetExitStatus(); code != 1 {
		t.Errorf("git %d, want %d", code, 1)
	}
}

// TestSetExitStatus tests that a call to SetExitStatus(n) sets the exit status
// to n.
func TestSetExitStatus(t *testing.T) {
	SetExitStatus(2)
	if code := GetExitStatus(); code != 2 {
		t.Errorf("git %d, want %d", code, 2)
	}
}

// Exit (and AtExit), ExitIfErrors and Fatalf can not be tested since they call
// os.Exit.
