/*
 * mon_lang - code_gen
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package codegen

import "fmt"

// Comment is an AsmInstruction that carries a human-readable note. The native
// encoder skips it (see encoder/program.go); it survives purely as an
// annotation in the asm AST.
type Comment struct {
	Text string
}

func (a Comment) Ir() string {
	return fmt.Sprintf("# %s", a.Text)
}
