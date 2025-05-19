package codegen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/util"
)

type ReplacementState struct {
	CurrentOffset int
	OffsetMap     map[string]int
}

type ReplacementPassGen struct {
}

func NewReplacementPassGen() ReplacementPassGen {
	return ReplacementPassGen{}
}

func (r *ReplacementPassGen) ReplaceOperand(operand AsmOperand, state ReplacementState) (ReplacementState, AsmOperand) {
	pseudo, isit := operand.(Pseudo)
	if isit {
		value, exists := state.OffsetMap[pseudo.Ident]
		if exists {
			return state, Stack{value}
		} else {
			state.CurrentOffset = state.CurrentOffset - 4
			state.OffsetMap[pseudo.Ident] = state.CurrentOffset
			return state, Stack{state.CurrentOffset}
		}
	} else {
		return state, operand
	}
}

func (r *ReplacementPassGen) ReplacePseudosInInstruction(instr AsmInstruction, state ReplacementState) (ReplacementState, AsmInstruction) {
	switch ast := instr.(type) {
	case DeallocateStack:
		return state, instr
	case Label:
		return state, instr
	case SetCC:
		replacedState, op := r.ReplaceOperand(ast.Op, state)
		return replacedState, SetCC{
			Op: op,
			CC: ast.CC,
		}
	case Cmp:
		fmt.Printf("[debug] Cmp before replacement - Type: %v\n", ast.Type)
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, Cmp{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
		}
	case AsmMov:
		fmt.Printf("[debug] AsmMov before replacement - Type: %v\n", ast.Type)
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmMov{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
		}
	case Unary:
		fmt.Printf("[debug] Unary before replacement - Type: %v\n", ast.Type)
		replacedState, dst := r.ReplaceOperand(ast.Dst, state)
		return replacedState, Unary{
			Type: ast.Type,
			Dst:  dst,
			Op:   ast.Op,
		}
	case AsmBinary:
		fmt.Printf("[debug] AsmBinary before replacement - Type: %v\n", ast.Type)
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmBinary{
			Type: ast.Type,
			Src:  src,
			Dst:  dst,
			Op:   ast.Op,
		}
	case Idiv:
		fmt.Printf("[debug] Idiv before replacement - Type: %v\n", ast.Type)
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		return replacedState, Idiv{
			Type: ast.Type,
			Src:  src,
		}
	case Return:
		return state, instr
	case Cdq:
		fmt.Printf("[debug] Cdq before replacement - Type: %v\n", ast.Type)
		return state, Cdq{
			Type: ast.Type,
		}
	case AllocateStack:
		panic("you are not belong to us")
	default:
		return state, instr
	}
}

func (r *ReplacementPassGen) ReplacePseudosInFn(fn AsmFnDef, state ReplacementState) (ReplacementState, AsmFnDef) {
	for i, instr := range fn.Irs {
		replacedState, replaced := r.ReplacePseudosInInstruction(instr, state)
		fn.Irs[i] = replaced
		state = replacedState
	}
	return state, fn
}

func (r *ReplacementPassGen) ReplacePseudosInProgram(program AsmProgram, symbolTable *symbols.SymbolTable) AsmProgram {
	initState := ReplacementState{
		CurrentOffset: 0,
		OffsetMap:     make(map[string]int),
	}
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		finalState, asmFnDef := r.ReplacePseudosInFn(fn, initState)
		asmFnDefs = append(asmFnDefs, asmFnDef)
		symbolTable.SetBytesRequired(fn.Ident, util.Abs(finalState.CurrentOffset))
	}
	return AsmProgram{AsmFnDef: asmFnDefs, AsmExternFn: program.AsmExternFn}
}
