// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"strings"
	"testing"
)

// TestCommandLongName tests the Command.LongName method.
func TestCommandLongName(t *testing.T) {
	var tests = []struct {
		names []string
		want  string
	}{
		{[]string{"test"}, ""},
		{[]string{"test", "cmd"}, "cmd"},
		{[]string{"test", "cmd", "a"}, "cmd a"},
		{[]string{"test", "cmd", "a", "b"}, "cmd a b"},
	}

	for _, test := range tests {
		t.Run(mkname(test.want), func(t *testing.T) {
			// Build the command and the subcommands.
			var (
				pcmd *Command
				cmd  *Command
			)
			for _, name := range test.names {
				cmd = &Command{Name: name}
				cmd.parent = pcmd
				pcmd = cmd
			}

			got := cmd.LongName()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// TestCommandString tests the Command.String method.
func TestCommandString(t *testing.T) {
	var tests = []struct {
		names []string
		want  string
	}{
		{[]string{"test"}, "test"},
		{[]string{"test", "cmd"}, "test cmd"},
		{[]string{"test", "cmd", "a"}, "test cmd a"},
		{[]string{"test", "cmd", "a", "b"}, "test cmd a b"},
	}

	for _, test := range tests {
		t.Run(mkname(test.want), func(t *testing.T) {
			// Build the command and the subcommands.
			var pcmd *Command
			cmd := &Command{Name: "test"}
			for _, name := range test.names {
				cmd = &Command{Name: name}
				cmd.parent = pcmd
				pcmd = cmd
			}

			got := cmd.String()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// mkname returns a suitable name to use for a sub test.
func mkname(s string) string {
	if s == "" {
		// Replace the empty string with MIDDLE DOT ('·').
		return "\u00B7"
	}

	// Replace the SPACE (' ') character with OPEN BOX ('␣').
	return strings.Replace(s, " ", "\u2423", -1)
}
