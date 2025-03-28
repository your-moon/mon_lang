package codegen

type ReplacementPassGen struct {
	CurrentOffset int
	OffsetMap     map[string]int
}

func NewReplacementPassGen() ReplacementPassGen {
	return ReplacementPassGen{
		CurrentOffset: 0,
		OffsetMap:     make(map[string]int),
	}
}

func (r *ReplacementPassGen) ReplaceOperand(operand AsmOperand) AsmOperand {
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

func (r *ReplacementPassGen) ReplacePseudosInInstruction(instr AsmInstruction) AsmInstruction {
	switch ast := instr.(type) {
	case SetCC:
		op := r.ReplaceOperand(ast.Op)
		return SetCC{
			Op: op,
			CC: ast.CC,
		}
	case Cmp:
		src := r.ReplaceOperand(ast.Src)
		dst := r.ReplaceOperand(ast.Dst)
		return Cmp{
			Src: src,
			Dst: dst,
		}
	case AsmMov:
		src := r.ReplaceOperand(ast.Src)
		dst := r.ReplaceOperand(ast.Dst)
		return AsmMov{
			Src: src,
			Dst: dst,
		}
	case Unary:
		dst := r.ReplaceOperand(ast.Dst)
		return Unary{
			Dst: dst,
			Op:  ast.Op,
		}
	case AsmBinary:
		src := r.ReplaceOperand(ast.Src)
		dst := r.ReplaceOperand(ast.Dst)
		return AsmBinary{
			Src: src,
			Dst: dst,
			Op:  ast.Op,
		}
	case Idiv:
		src := r.ReplaceOperand(ast.Src)
		return Idiv{
			Src: src,
		}
	case Return:
		return instr
	case Cdq:
		return instr
	case AllocateStack:
		panic("you are not belong to us")
	default:
		return instr
		// panic("unimplemented pseudo ininstruction")
	}
}

func (r *ReplacementPassGen) ReplacePseudosInFn(fn AsmFnDef) AsmFnDef {
	for i, instr := range fn.Irs {
		replaced := r.ReplacePseudosInInstruction(instr)
		fn.Irs[i] = replaced
	}
	return fn
}

func (r *ReplacementPassGen) ReplacePseudosInProgram(program AsmProgram) AsmProgram {
	program.AsmFnDef = r.ReplacePseudosInFn(program.AsmFnDef)
	return program
}
