package codegen

type ReplacementGen struct {
	CurrentOffset int
	OffsetMap     map[string]int
}

func NewReplacement() ReplacementGen {
	return ReplacementGen{
		CurrentOffset: 0,
		OffsetMap:     make(map[string]int),
	}
}

func (r *ReplacementGen) ReplaceOperand(operand AsmOperand) AsmOperand {
	pseudo, isit := operand.(Pseudo)
	if isit {
		value, exists := r.OffsetMap[pseudo.Ident]
		if exists {
			return Stack{value}
		} else {
			r.CurrentOffset = r.CurrentOffset - 4
			r.OffsetMap[pseudo.Ident] = r.CurrentOffset
			return Stack{r.CurrentOffset}
		}
	} else {
		return operand
	}
}

func (r *ReplacementGen) ReplacePseudosInInstruction(instr AsmInstruction) AsmInstruction {
	switch ast := instr.(type) {
	case Mov:
		src := r.ReplaceOperand(ast.Src)
		dst := r.ReplaceOperand(ast.Dst)
		return Mov{
			Src: src,
			Dst: dst,
		}
	case Unary:
		dst := r.ReplaceOperand(ast.Dst)
		return Unary{
			Dst: dst,
			Op:  ast.Op,
		}
	case Return:
		return instr
	case AllocateStack:
		panic("you are not belong to us")
	default:
		panic("unimplemented pseudo ininstruction")

	}
}

func (r *ReplacementGen) ReplacePseudosInFn(fn AsmFnDef) AsmFnDef {
	for i, instr := range fn.Irs {
		replaced := r.ReplacePseudosInInstruction(instr)
		fn.Irs[i] = replaced
	}
	return fn
}

func (r *ReplacementGen) ReplacePseudosInProgram(program AsmProgram) AsmProgram {
	program.AsmFnDef = r.ReplacePseudosInFn(program.AsmFnDef)
	return program
}
