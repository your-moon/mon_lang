package codegen

import (
	"github.com/your-moon/mn_compiler_go_version/code_gen/asmsymbol"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/util"
	"github.com/your-moon/mn_compiler_go_version/util/roundingutil"
)

type ReplacementState struct {
	CurrentOffset int
	OffsetMap     map[string]int
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
