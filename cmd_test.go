// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"strings"
	"testing"
)

type list = []string

// TestCommandLongName tests the Command.LongName method.
func TestCommandLongName(t *testing.T) {
	var tests = []struct {
		names list
		want  string
	}{
		{list{"test"}, ""},
		{list{"test", "cmd"}, "cmd"},
		{list{"test", "cmd", "a"}, "cmd a"},
		{list{"test", "cmd", "a", "b"}, "cmd a b"},
	}

	for _, test := range tests {
		t.Run(mkname(test.want), func(t *testing.T) {
			cmd := buildp(test.names)
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
		names list
		want  string
	}{
		{list{"test"}, "test"},
		{list{"test", "cmd"}, "test cmd"},
		{list{"test", "cmd", "a"}, "test cmd a"},
		{list{"test", "cmd", "a", "b"}, "test cmd a b"},
	}

	for _, test := range tests {
		t.Run(mkname(test.want), func(t *testing.T) {
			cmd := buildp(test.names)
			got := cmd.String()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// buildp returns a command tree, with the parent field set correctly.
func buildp(tree []string) *Command {
	var parent, cmd *Command
	for _, name := range tree {
		cmd = &Command{Name: name}
		cmd.parent = parent
		parent = cmd
	}

	return cmd
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
