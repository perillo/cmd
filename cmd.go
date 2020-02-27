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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

// ErrNoCommand is returned by Parse when no command was provided.
var ErrNoCommand = errors.New("no command")

// ErrUnknowCommand is returned by Parse when an unknown command was provided.
var ErrUnknownCommand = errors.New("unknown command")

// A Command is an implementation of a single command.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageFunc will replace Command.Usage, if specified.
	UsageFunc func()

	// Name is the command name
	Name string

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

	// parent is the parent of this command.
	parent *Command
}

// LongName returns the command's long name.
func (c *Command) LongName() string {
	if c.parent == nil {
		return "" // avoid panic if called on the main command
	}

	name := c.Name
	for cmd := c.parent; cmd.parent != nil; cmd = cmd.parent {
		name = cmd.Name + " " + name
	}

	return name
}

// Runnable reports whether the command can be run.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// defaultUsage prints a usage message documenting all defined command-line
// flags to os.Stderr.
func (c *Command) defaultUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	c.Flag.PrintDefaults()
	if c.Long != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", c.Long)
	}

	if len(c.Commands) > 0 {
		fmt.Fprint(os.Stderr, "\ncommands:\n\n")
		for _, cmd := range c.Commands {
			fmt.Fprintf(os.Stderr, "\t%-11s %s\n", cmd.Name, cmd.Short)
		}
	}
}

func (c *Command) Usage() {
	if c.UsageFunc != nil {
		c.UsageFunc()

		return
	}

	c.defaultUsage()
}

// Parse parses command-line from argument list, which should not include
// the command name, and return the selected Command and arguments.  Must be
// called after all flags in cmdline are defined and before flags are accessed
// by the program.  The return value will be ErrHelp if -help or -h were set
// but not defined.
func Parse(main *Command, cmdline *flag.FlagSet, argv []string) (
	*Command, []string, error) {

	// Configure cmdline so that errors and output are in our control.
	cmdline.Init("", flag.ContinueOnError)
	cmdline.Usage = func() {}
	cmdline.SetOutput(ioutil.Discard)
	if err := cmdline.Parse(argv); err != nil {
		return main, nil, err
	}

	args := cmdline.Args()
	if len(args) < 1 {
		return main, args, ErrNoCommand
	}

	// TODO(mperillo): Sub commands are not currently supported.
	for _, cmd := range Main.Commands {
		if cmd.Name != args[0] {
			continue
		}
		if !cmd.Runnable() {
			continue
		}
		cmd.parent = main

		// Configure cmd.Flag as it was done with cmdline, but restore the
		// output when returning, since cmd.Usage requires it.
		cmd.Flag.Init("", flag.ContinueOnError)
		cmd.Flag.Usage = func() {}
		defer disable(&cmd.Flag)()
		if cmd.CustomFlags {
			args = args[1:]
		} else {
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				return cmd, nil, err
			}
			args = cmd.Flag.Args()
		}

		return cmd, args, nil
	}

	return main, cmdline.Args(), ErrUnknownCommand
}

func disable(f *flag.FlagSet) (enable func()) {
	w := f.Output()
	f.SetOutput(ioutil.Discard)

	return func() {
		f.SetOutput(w)
	}
}

// Run parses the command-line from os.Args[1:] and execute the appropriate
// sub command of the Main command.
func Run() {
	cmd, args, err := Parse(Main, flag.CommandLine, os.Args[1:])
	switch {
	case err == ErrUnknownCommand:
		fmt.Fprintf(os.Stderr, "%s %s: unknown command\n", Main.Name, args[0])
		fmt.Fprintf(os.Stderr, "Run '%s -help' for usage.\n", Main.Name)
	case err == flag.ErrHelp:
		cmd.Usage()
	case err != nil:
		fmt.Fprintf(os.Stderr, "%s: %v\n", Main.Name, err)
		cmd.Usage()
	}
	if err != nil {
		SetExitStatus(2)
		Exit()
	}

	cmd.Run(cmd, args)
	Exit()
}

// Main is the main command.
//
// Name, UsageLine and Long fields should be set by the user.
var Main = &Command{}
