/*
 * mon_lang - mon-lsp
 *
 * Language server for mon_lang, speaking LSP over stdio.
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package main

import (
	"fmt"
	"os"

	"github.com/your-moon/mon_lang/lsp"
)

func main() {
	server := lsp.NewServer(os.Stdin, os.Stdout)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mon-lsp: %v\n", err)
		os.Exit(1)
	}
}
