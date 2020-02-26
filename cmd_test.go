// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"testing"
)

// TestLongName tests the Command.LongName method.
func TestLongName(t *testing.T) {
	var tests = []struct {
		name  string
		usage string
		want  string
	}{
		{"test", "test cmd", "cmd"},
		{"test", "test cmd a", "cmd a"},
	}

	for _, test := range tests {
		t.Run(test.usage, func(t *testing.T) {
			Name = test.name
			cmd := &Command{UsageLine: test.usage}
			got := cmd.LongName()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// TestName tests the Command.Name method.
func TestName(t *testing.T) {
	var tests = []struct {
		name  string
		usage string
		want  string
	}{
		{"test", "test cmd", "cmd"},
		{"test", "test cmd a", "a"},
	}

	for _, test := range tests {
		t.Run(test.usage, func(t *testing.T) {
			Name = test.name
			cmd := &Command{UsageLine: test.usage}
			got := cmd.Name()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
