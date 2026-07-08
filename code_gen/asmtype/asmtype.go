/*
 * mon_lang - code_gen/asmtype
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package asmtype

type AsmType interface {
	asmtype()
}

type LongWord struct{}

func (l *LongWord) asmtype() {}

type QuadWord struct{}

func (l *QuadWord) asmtype() {}

type StringType struct{}

func (s *StringType) asmtype() {}

// Double is a 64-bit IEEE-754 float living in XMM registers.
type Double struct{}

func (d *Double) asmtype() {}
