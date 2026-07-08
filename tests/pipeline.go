/*
 * mon_lang - test harness pipeline glue
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package tests

import (
	"bytes"

	codegen "github.com/your-moon/mon_lang/code_gen"
	"github.com/your-moon/mon_lang/linker"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/tackygen"
	"github.com/your-moon/mon_lang/util"
	"github.com/your-moon/mon_lang/util/unique"
)

func tackygenNew(u unique.UniqueGen, table *symbols.SymbolTable) tackygen.TackyGen {
	return tackygen.NewTackyGen(u, table)
}

func tackygenOptimize(p tackygen.TackyProgram) tackygen.TackyProgram {
	return tackygen.Optimize(p)
}

// compileCC links through the legacy external as/cc path for the
// differential mode (MON_TEST_CC=1).
func compileCC(lnk *linker.Linker, prog codegen.AsmProgram) error {
	buf := new(bytes.Buffer)
	w := codegen.NewGenASM(buf, util.GetOsType())
	w.GenAsm(prog)
	lnk.SetAssemblyContent(buf.String())
	if err := lnk.Link(); err != nil {
		return err
	}
	return lnk.MakeExecutable()
}
