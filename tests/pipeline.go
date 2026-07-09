/*
 * mon_lang - test harness pipeline glue
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package tests

import (
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/tackygen"
	"github.com/your-moon/mon_lang/util/unique"
)

func tackygenNew(u unique.UniqueGen, table *symbols.SymbolTable) tackygen.TackyGen {
	return tackygen.NewTackyGen(u, table)
}

func tackygenOptimize(p tackygen.TackyProgram) tackygen.TackyProgram {
	return tackygen.Optimize(p)
}
