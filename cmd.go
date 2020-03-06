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

// Standard Posix exit status constants.
const (
	ExitSuccess = iota
	ExitFailure
	ExitUsageError
)

// ErrHelp is the error returned by Parse if the -help or -h flag is invoked
// but no such flag is defined.
var ErrHelp = flag.ErrHelp

// ErrNoCommand is the error returned by Parse when no command is invoked.
var ErrNoCommand = errors.New("no command")

// ErrUnknownCommand is the error returned by Parse when an unknown command is
// invoked.
var ErrUnknownCommand = errors.New("unknown command")

// A Command is an implementation of a single command.
type Command struct {
	// Run runs the command and returns the exit status.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string) int

	// Usage prints the command usage to os.Stderr.  If not specified a default
	// template will be used, printing UsageLine, followed by a call to
	// Flag.PrintDefaults and a list of available sub commands.
	Usage func()

	// Name is the command name.
	Name string

	// UsageLine is the one-line usage message.  The message must not contain
	// the command name, since it will be added automatically in the default
	// usage template.
	UsageLine string

	// Short is the short description shown in the 'cmd -help' output.
	Short string

	// Long is the long message shown in the command default usage output.
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

// String implements the Stringer interface.
func (c *Command) String() string {
	// Return the full name of the command.
	name := c.Name
	for cmd := c.parent; cmd != nil; cmd = cmd.parent {
		name = cmd.Name + " " + name
	}

	return name
}

// defaultUsage prints a usage message documenting all defined command-line
// flags and sub commands to os.Stderr.
func (c *Command) defaultUsage() {
	printf("usage: %s %s\n", c, c.UsageLine)
	c.Flag.PrintDefaults()
	if c.Long != "" {
		printf("\n%s\n", c.Long)
	}

	if len(c.Commands) > 0 {
		print("\ncommands:\n\n")
		for _, cmd := range c.Commands {
			printf("\t%-11s %s\n", cmd.Name, cmd.Short)
		}
	}
}

func (c *Command) usage() {
	if c.Usage != nil {
		c.Usage()

		return
	}

	c.defaultUsage()
}

// Parse parses command-line from argument list, which should not include the
// main command name, and return the invoked Command.
//
// Parse must be called after all flags in main commands Flag are defined and
// before flags are accessed by the program.  The return value will be
// flag.ErrHelp if -help or -h were set but not defined.
func Parse(main *Command, argv []string) (*Command, error) {
	// Configure main.Flag so that errors and output are in our control, but
	// restore the output when returning, since Command.defaultUsage will
	// require it.
	defer configure(main)()
	if err := main.Flag.Parse(argv); err != nil {
		return main, err
	}

	args := main.Flag.Args()
	if len(args) < 1 {
		return main, ErrNoCommand
	}

	// TODO(mperillo): Sub commands are not currently supported.
	for _, cmd := range main.Commands {
		if cmd.Name != args[0] {
			continue
		}
		cmd.parent = main

		// Configure cmd.Flag as it was done with main.Flag.
		defer configure(cmd)()
		if cmd.CustomFlags {
			args = args[1:]
		} else {
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				return cmd, err
			}
		}

		return cmd, nil
	}

	return main, ErrUnknownCommand
}

// configure configures c so that c.Flag error handling is set to continue on
// errors and its output is temporarily disabled.  Calling the returned restore
// function will restore C.Flag.Output to os.Stderr and set c.Flag.Usage to
// c.usage.
//
// configure assumes that c.Flag has not been modified, so that c.Flag.Output()
// is os.Stderr and c.Flag.Usage is nil.
func configure(c *Command) (restore func()) {
	c.Flag.Init(c.String(), flag.ContinueOnError)
	c.Flag.SetOutput(ioutil.Discard)

	return func() {
		c.Flag.Usage = c.usage // this is not really necessary
		c.Flag.SetOutput(nil)
	}
}

func print(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// Run parses the command-line from os.Args[1:] and execute the appropriate
// sub command of main.  It returns the status code returned by Command.Run or
// ExitUsageError in case of parsing error.
func Run(main *Command) int {
	cmd, err := Parse(main, os.Args[1:])
	osname := os.Args[0] // follow UNIX cmd -h convention
	args := cmd.Flag.Args()
	switch {
	case err == ErrUnknownCommand:
		main.Name = osname
		printf("%s %s: unknown command\n", cmd, args[0])
		printf("Run '%s -help' for usage.\n", cmd)
	case err == flag.ErrHelp:
		main.Name = osname
		cmd.usage()
	case err != nil:
		main.Name = osname
		printf("%s: %v\n", cmd, err)
		cmd.usage()
	}
	if err != nil {
		return ExitUsageError
	}
	if !cmd.Runnable() {
		printf("%s: not runnable\n", cmd)

		return ExitUsageError
	}

	return cmd.Run(cmd, args)
}
