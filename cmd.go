// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The implementation of the cmd package has been adapted from
// src/cmd/go/internal/base/base.go and src/cmd/go/main.go
// in the Go source distribution.
// Copyright 2017 The Go Authors. All rights reserved.

// Package cmd implements a simple way for a single command to have many
// subcommands.
package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// A Command is an implementation of a single command.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageFunc will replace Command.Usage, if specified.
	UsageFunc func()

	// UsageLine is the one-line usage message.
	UsageLine string

	// Short is the short description shown in the 'cmd -help' output.
	Short string

	// Long is the long message shown in the Command.Usage output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own flag parsing.
	CustomFlags bool

	// Commands lists the available commands.
	// The order here is the order in which they are printed by 'cmd -help'.
	// Note that subcommands are in general best avoided.
	Commands []*Command
}

// LongName returns the command's long name: all the words in the usage line
// between cmd Name and a flag or argument,
func (c *Command) LongName() string {
	name := c.UsageLine
	if i := strings.Index(name, " ["); i >= 0 {
		name = name[:i]
	}
	if name == Name {
		return ""
	}

	return strings.TrimPrefix(name, Name+" ")
}

// Name returns the command's short name: the last word in the usage line
// before a flag or argument.
func (c *Command) Name() string {
	name := c.LongName()
	if i := strings.LastIndex(name, " "); i >= 0 {
		name = name[i+1:]
	}

	return name
}

// Runnable reports whether the command can be run.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// defaultUsage prints a usage message documenting all defined command-line
// flags to os.Stderr and exit with an exit status 2.
func (c *Command) defaultUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	c.Flag.PrintDefaults()
	if c.Long != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", c.Long)
	}

	if len(c.Commands) > 0 {
		fmt.Fprint(os.Stderr, "\ncommands:\n\n")
		for _, cmd := range c.Commands {
			fmt.Fprintf(os.Stderr, "\t%s    %s\n", cmd.LongName(), cmd.Short)
		}
	}
}

func (c *Command) Usage() {
	if c.UsageFunc != nil {
		c.UsageFunc()

		return
	}

	c.defaultUsage()

	// Exit as a convenience in case Usage is called by the user, instead of
	// the flag package.
	SetExitStatus(2)
	Exit()
}

// Run parses the command-line from os.Args[1:] and execute the appropriate
// sub command of the Main command.
func Run() {
	flag.Usage = Main.Usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		Main.Usage()

		return // in case Main.Usage does not call Exit
	}

	// TODO(mperillo): Sub commands are not currently supported.
	for _, cmd := range Main.Commands {
		if cmd.Name() != args[0] {
			continue
		}
		if !cmd.Runnable() {
			continue
		}

		// Initialize cmd.Flag to have the same default error handling as
		// flag.CommandLine, in case cmd.UsageFunc is specified and it does
		// not call Exit.
		cmd.Flag.Init("", flag.ExitOnError)
		cmd.Flag.Usage = func() { cmd.Usage() }
		if cmd.CustomFlags {
			args = args[1:]
		} else {
			cmd.Flag.Parse(args[1:]) // will call os.Exit(2) in case of errors
			args = cmd.Flag.Args()
		}

		cmd.Run(cmd, args)
		Exit()

		return
	}

	fmt.Fprintf(os.Stderr, "%s %s: unknown command\n", Name, args[0])
	fmt.Fprintf(os.Stderr, "Run '%s -help' for usage.\n", Name)
	SetExitStatus(2)
	Exit()
}

// Name is the main command name.
var Name string

// Main is the main command.
//
// UsageLine and Long fields should be set by the user.
var Main = &Command{}
