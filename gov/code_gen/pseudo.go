package codegen

import semanticanalysis "github.com/your-moon/mn_compiler_go_version/semantic_analysis"

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
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, Cmp{
			Src: src,
			Dst: dst,
		}
	case AsmMov:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmMov{
			Src: src,
			Dst: dst,
		}
	case Unary:
		replacedState, dst := r.ReplaceOperand(ast.Dst, state)
		return replacedState, Unary{
			Dst: dst,
			Op:  ast.Op,
		}
	case AsmBinary:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		replacedState, dst := r.ReplaceOperand(ast.Dst, replacedState)
		return replacedState, AsmBinary{
			Src: src,
			Dst: dst,
			Op:  ast.Op,
		}
	case Idiv:
		replacedState, src := r.ReplaceOperand(ast.Src, state)
		return replacedState, Idiv{
			Src: src,
		}
	case Return:
		return state, instr
	case Cdq:
		return state, instr
	case AllocateStack:
		panic("you are not belong to us")
	default:
		return state, instr
		// panic("unimplemented pseudo ininstruction")
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

func (r *ReplacementPassGen) ReplacePseudosInProgram(program AsmProgram, symbolTable *semanticanalysis.SymbolTable) AsmProgram {
	initState := ReplacementState{
		CurrentOffset: 0,
		OffsetMap:     make(map[string]int),
	}
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		finalState, asmFnDef := r.ReplacePseudosInFn(fn, initState)
		asmFnDefs = append(asmFnDefs, asmFnDef)
		symbolTable.SetBytesRequired(fn.Ident, finalState.CurrentOffset)
	}
	return AsmProgram{AsmFnDef: asmFnDefs}
}
