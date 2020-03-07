// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"reflect"
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

// TestParse tests the Parse function.
func TestParse(t *testing.T) {
	// Define variables to keep the test entries short.
	cmd0 := list{"test"}
	cmd1 := list{"test", "cmd"}
	cmd2 := list{"test", "cmd1", "cmd2"}

	var tests = []struct {
		names list
		argv  list
		cmd   string // expected command name
		err   error  // expected error
	}{
		{cmd0, list{"test"}, "test", ErrNoCommand},
		{cmd0, list{"test", "-h"}, "test", ErrHelp},
		{cmd1, list{"test", "cmd"}, "cmd", nil},
		{cmd1, list{"test", "a"}, "test", ErrUnknownCommand},
		{cmd1, list{"test", "cmd", "-h"}, "cmd", ErrHelp},

		{cmd2, list{"test", "cmd1"}, "cmd1", ErrNoCommand},
		{cmd2, list{"test", "cmd1", "cmd2"}, "cmd2", nil},
		{cmd2, list{"test", "cmd1", "a"}, "cmd1", ErrUnknownCommand},
		{cmd2, list{"test", "cmd1", "cmd2", "-h"}, "cmd2", ErrHelp},
	}

	for _, test := range tests {
		name := join(test.names) + ":" + join(test.argv)
		t.Run(mkname(name), func(t *testing.T) {
			main := build(test.names)

			cmd, err := Parse(main, test.argv[1:])
			if err != test.err {
				t.Errorf("got error %v, want %v", err, test.err)
			}
			if cmd.Name != test.cmd {
				t.Errorf("got command %q, want %q", cmd.Name, test.cmd)
			}
		})
	}
}

// TestParseFlag tests the Parse function, with a flag and an argument.
func TestParseFlag(t *testing.T) {
	// Define variables to keep the test entries short.
	cmd1 := list{"test", "cmd"}
	cmd2 := list{"test", "cmd1", "cmd2"}

	var tests = []struct {
		names list
		argv  list
		depth int
	}{
		{cmd1, list{"test", "cmd", "-flag", "arg"}, 1},
		{cmd2, list{"test", "cmd1", "cmd2", "-flag", "arg"}, 2},
	}

	for _, test := range tests {
		name := join(test.names) + ":" + join(test.argv)
		t.Run(mkname(name), func(t *testing.T) {
			main := build(test.names)
			flag := find(main, test.depth).Flag.Bool("flag", false, "flag")

			cmd, err := Parse(main, test.argv[1:])
			arg := cmd.Flag.Arg(0)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if !*flag {
				t.Errorf("flag not set")
			}
			if arg != "arg" {
				t.Errorf("got argument %q, want %q", arg, "arg")
			}
		})
	}
}

// TestParseCustomFlags tests the Parse function, when the command has the
// CustomFlags field set to true..
func TestParseCustomFlags(t *testing.T) {
	tree := list{"test", "cmd"}
	argv := list{"test", "cmd", "-flag", "arg"}

	main := build(tree)
	main.Commands[0].CustomFlags = true
	flag := main.Commands[0].Flag.Bool("flag", false, "flag")

	cmd, err := Parse(main, argv[1:])
	args := cmd.Flag.Args()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	// Test that the sub-command flag has not been parsed.
	if *flag {
		t.Errorf("flag set")
	}
	if !reflect.DeepEqual(args, argv[2:]) {
		t.Errorf("got arguments %q, want %q", args, argv[2:])
	}
}

// TestParseMainFlagsSet tests the Parse function, when the main command has
// flags set and additional sub commands.
func TestParseMainFlagsSet(t *testing.T) {
	tree := list{"test", "cmd"}
	argv := list{"test", "-flag0", "cmd", "-flag1", "arg"}

	main := build(tree)
	flag0 := main.Flag.Bool("flag0", false, "flag0")
	flag1 := main.Commands[0].Flag.Bool("flag1", false, "flag1")

	cmd, err := Parse(main, argv[1:])
	arg := cmd.Flag.Arg(0)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	// Test that the invoked command is the sub-command, and not the main
	// command.
	if cmd.Name != "cmd" {
		t.Errorf("got command %q, want %q", cmd.Name, "cmd")
	}
	if !*flag0 {
		t.Errorf("flag0 not set")
	}
	if !*flag1 {
		t.Errorf("flag1 not set")
	}
	if arg != "arg" {
		t.Errorf("got argument %q, want %q", arg, "arg")
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

// build returns a command tree, with the Commands field set correctly.
func build(tree []string) *Command {
	// Traverse the tree in reverse order.
	var child, cmd *Command
	for i := len(tree) - 1; i >= 0; i-- {
		cmd = &Command{Name: tree[i]}
		if child != nil {
			cmd.Commands = []*Command{child}
		}
		child = cmd
	}

	return cmd
}

// find finds the sub-command of main at the specified depth.  A depth 0 will
// return main.
func find(main *Command, depth int) *Command {
	cmd := main
	for i := 0; i < depth; i++ {
		cmd = cmd.Commands[0]
	}

	return cmd
}

func join(elems []string) string {
	return strings.Join(elems, " ")
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
