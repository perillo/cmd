// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd_test

import (
	"fmt"
	"os"

	"github.com/perillo/cmd"
)

var (
	main = &cmd.Command{
		Name:      "example",
		UsageLine: "<command> [arguments]",
		Long:      "example is an example command",
	}

	child = &cmd.Command{
		Name:      "child",
		UsageLine: "[-v]",
		Short:     "sub command example",
	}
)

func init() {
	child.Run = runChild // break initialization loop
}

var verbose = child.Flag.Bool("v", false, "verbose")

func Example() {
	// We are going to modify os.Args; make sure to restore it.
	defer restore()

	// Setup main commands.
	main.Commands = []*cmd.Command{child}

	// Call Run with a custom os.Args.
	os.Args = []string{"example", "child", "-v", "a", "b"}
	status := cmd.Run(main)
	fmt.Println("exit status:", status)

	// Output:
	// full command name: "example child"
	// -v flag: true
	// args: [a b]
	// exit status: 0
}

func runChild(c *cmd.Command, args []string) int {
	fmt.Printf("full command name: %q\n", c)
	fmt.Println("-v flag:", *verbose)
	fmt.Println("args:", args)

	return cmd.ExitSuccess
}

// restore returns a function that, when called, will restore the global state
// modified during a test.
func restore() func() {
	args := os.Args

	return func() {
		os.Args = args
	}
}
