/*
 * mon_lang - code_gen
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package codegen

import (
	"sort"

	"github.com/your-moon/mon_lang/code_gen/asmsymbol"
	"github.com/your-moon/mon_lang/code_gen/asmtype"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util"
	"github.com/your-moon/mon_lang/util/roundingutil"
)

type ReplacementState struct {
	CurrentOffset int
	OffsetMap     map[string]int
	RegMap        map[string]AsmRegister // pseudos assigned to callee-saved regs
}

type ReplacementPassGen struct {
	asmSymbol *asmsymbol.SymbolTable
}

func NewReplacementPassGen(table *asmsymbol.SymbolTable) ReplacementPassGen {
	return ReplacementPassGen{
		asmSymbol: table,
	}
}

func (r *ReplacementPassGen) ReplaceOperand(operand AsmOperand, state ReplacementState) (ReplacementState, AsmOperand) {
	pseudo, isPseudo := operand.(Pseudo)

	if isPseudo {
		if reg, ok := state.RegMap[pseudo.Ident]; ok {
			return state, Register{Reg: reg}
		}
		value, exists := state.OffsetMap[pseudo.Ident]
		if exists {
			return state, Stack{value}
		} else {
			offsetSize, err := r.asmSymbol.GetSize(pseudo.Ident)
			if err != nil {
				panic(err)
			}
			offsetAlignment, err := r.asmSymbol.GetAlignment(pseudo.Ident)
			if err != nil {
				panic(err)
			}

			newOffset := roundingutil.RoundAwayFromZero(offsetAlignment, state.CurrentOffset-offsetSize)
			state.CurrentOffset = newOffset
			state.OffsetMap[pseudo.Ident] = state.CurrentOffset
			return state, Stack{state.CurrentOffset}
		}
	} else {
		return state, operand
	}

}

func (r *ReplacementPassGen) ReplacePseudosInInstruction(instr AsmInstruction, state ReplacementState) (ReplacementState, AsmInstruction) {
	switch ast := instr.(type) {
	case Label:
		return state, instr
	case SetCC:
		replacedState, op := r.ReplaceOperand(ast.Op, state)
		return replacedState, SetCC{
			Op: op,
			CC: ast.CC,
		}
	case Cmp:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, Cmp{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
		}
	case AsmMov:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmMov{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
		}
	case Unary:
		replacedState, dst := r.ReplaceOperand(ast.Dst, state)
		return replacedState, Unary{
			Type: ast.Type,
			Dst:  dst,
			Op:   ast.Op,
		}
	case AsmBinary:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmBinary{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
			Op:   ast.Op,
		}
	case Idiv:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		return replacedState, Idiv{
			Type: ast.Type,
			Src:  src,
		}
	case AsmMovSx:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmMovSx{
			Src: src,
			Dst: dst,
		}
	case AsmLea:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmLea{
			Src: src,
			Dst: dst,
		}
	case AsmCvtSi2Sd:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmCvtSi2Sd{Src: src, Dst: dst}
	case AsmCvtTsd2Si:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmCvtTsd2Si{Src: src, Dst: dst}
	case AsmLoadFromMem:
		replacedState, dst := r.ReplaceOperand(ast.Dst, state)
		return replacedState, AsmLoadFromMem{
			Type: ast.Type,
			Base: ast.Base,
			Dst:  dst,
		}
	case AsmStoreToMem:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		return replacedState, AsmStoreToMem{
			Type: ast.Type,
			Src:  src,
			Base: ast.Base,
		}
	case Return:
		return state, instr
	case Cdq:
		return state, Cdq{
			Type: ast.Type,
		}
	case StringLiteral:
		return state, ast
	case AllocateStack:
		panic("you are not belong to us")
	default:
		return state, instr
	}
}

// calleeSaved is the pool the allocator draws from; each register is
// preserved across calls by the callee contract, so a value assigned to one
// survives the whole function with no live-range analysis needed.
var calleeSaved = []AsmRegister{BX, R12, R13, R14, R15}

// collectPseudoUses counts how often each non-double pseudo appears; double
// pseudos are excluded (they need XMM registers, not GPRs).
func (r *ReplacementPassGen) collectPseudoUses(fn AsmFnDef) (map[string]int, map[string]bool) {
	uses := map[string]int{}
	pinned := map[string]bool{} // must stay in memory (address is taken)
	var count func(op AsmOperand)
	count = func(op AsmOperand) {
		if p, ok := op.(Pseudo); ok && !r.asmSymbol.IsDouble(p.Ident) {
			uses[p.Ident]++
		}
	}
	pin := func(op AsmOperand) {
		if p, ok := op.(Pseudo); ok {
			pinned[p.Ident] = true
		}
	}
	for _, instr := range fn.Irs {
		switch a := instr.(type) {
		case AsmMov:
			count(a.Src)
			count(a.Dst)
		case AsmBinary:
			count(a.Src)
			count(a.Dst)
		case Cmp:
			count(a.Src)
			count(a.Dst)
		case Unary:
			count(a.Dst)
		case Idiv:
			count(a.Src)
		case SetCC:
			// setcc writes a byte; keep its target in memory so the text
			// emitter needn't render byte-register names
			pin(a.Op)
			count(a.Op)
		case AsmMovSx:
			count(a.Src)
			count(a.Dst)
		case AsmLea:
			// lea needs a memory source: the address-taken value can't
			// live in a register
			pin(a.Src)
			count(a.Src)
			count(a.Dst)
		case Push:
			count(a.Op)
		}
	}
	return uses, pinned
}

func (r *ReplacementPassGen) ReplacePseudosInFn(fn AsmFnDef, state ReplacementState) (ReplacementState, AsmFnDef) {
	// linear-scan-lite: give each of the hottest non-double pseudos a
	// dedicated callee-saved register for the whole function. No interval
	// overlap is possible since each register belongs to one pseudo.
	uses, pinned := r.collectPseudoUses(fn)
	type uc struct {
		name string
		n    int
	}
	ranked := make([]uc, 0, len(uses))
	for name, n := range uses {
		if n >= 2 && !pinned[name] { // skip singles and address-taken vars
			ranked = append(ranked, uc{name, n})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].n != ranked[j].n {
			return ranked[i].n > ranked[j].n
		}
		return ranked[i].name < ranked[j].name // deterministic
	})

	state.RegMap = map[string]AsmRegister{}
	var usedRegs []AsmRegister
	saveSlots := map[AsmRegister]int{}
	for i, u := range ranked {
		if i >= len(calleeSaved) {
			break
		}
		reg := calleeSaved[i]
		state.RegMap[u.name] = reg
		usedRegs = append(usedRegs, reg)
		// reserve an 8-byte stack slot to preserve the caller's value
		state.CurrentOffset -= 8
		saveSlots[reg] = state.CurrentOffset
	}

	for i, instr := range fn.Irs {
		replacedState, replaced := r.ReplacePseudosInInstruction(instr, state)
		fn.Irs[i] = replaced
		state = replacedState
	}

	// preserve the callee-saved registers we used: save on entry, restore
	// before every Return. Rendered as ordinary movs, so both backends and
	// the standard epilogue are untouched.
	if len(usedRegs) > 0 {
		saves := make([]AsmInstruction, 0, len(usedRegs))
		for _, reg := range usedRegs {
			saves = append(saves, AsmMov{Type: &asmtype.QuadWord{},
				Src: Register{Reg: reg}, Dst: Stack{Value: saveSlots[reg]}})
		}
		out := make([]AsmInstruction, 0, len(fn.Irs)+len(usedRegs)*2)
		out = append(out, saves...)
		for _, instr := range fn.Irs {
			if _, isRet := instr.(Return); isRet {
				for _, reg := range usedRegs {
					out = append(out, AsmMov{Type: &asmtype.QuadWord{},
						Src: Stack{Value: saveSlots[reg]}, Dst: Register{Reg: reg}})
				}
			}
			out = append(out, instr)
		}
		fn.Irs = out
	}

	return state, fn
}

func (r *ReplacementPassGen) ReplacePseudosInProgram(program AsmProgram, symbolTable *symbols.SymbolTable) AsmProgram {
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		// fresh slot map per function: names are globally unique today, but
		// stack slots must never alias across frames regardless
		initState := ReplacementState{
			CurrentOffset: 0,
			OffsetMap:     make(map[string]int),
		}
		finalState, asmFnDef := r.ReplacePseudosInFn(fn, initState)
		asmFnDefs = append(asmFnDefs, asmFnDef)
		symbolTable.SetBytesRequired(fn.Ident, util.Abs(finalState.CurrentOffset))
	}
	return AsmProgram{AsmFnDef: asmFnDefs, AsmExternFn: program.AsmExternFn, GlobalVars: program.GlobalVars}
}
