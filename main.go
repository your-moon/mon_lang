/*
 * mon_lang - compiler driver
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/your-moon/mon_lang/cli"
	"github.com/your-moon/mon_lang/linker"
)

func main() {
	compiler := cli.New()
	if err := compiler.Run(os.Args); err != nil {
		// a program run via --run owns its exit code; pass it through
		var exitErr *linker.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
