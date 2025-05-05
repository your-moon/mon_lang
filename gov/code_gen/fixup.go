package codegen

type FixUpPassGen struct {
	stackSize         int
	fixedInstructions []AsmInstruction
}

func NewFixUpPassGen(stackSize int) FixUpPassGen {
	return FixUpPassGen{
		stackSize:         stackSize,
		fixedInstructions: []AsmInstruction{},
	}
}

func (f *FixUpPassGen) AppendFixedInstruction(instr AsmInstruction) {
	f.fixedInstructions = append(f.fixedInstructions, instr)
}

// in golang if you give array to param, its call by ref that means instructions is mutable
func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction, instructions []AsmInstruction, idx int) {
	switch ast := instr.(type) {
	case Cmp:
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)

		// is invalid mov instruction
		if isit && isitdst {
			mov1 := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			cmp := Cmp{
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			f.AppendFixedInstruction(mov1)
			f.AppendFixedInstruction(cmp)
			return
		}

		dstimm, isitdst := ast.Dst.(Imm)
		if isitdst {
			mov1 := AsmMov{
				Src: dstimm,
				Dst: Register{Reg: R11},
			}
			cmp := Cmp{
				Src: ast.Src,
				Dst: Register{Reg: R11},
			}
			f.AppendFixedInstruction(mov1)
			f.AppendFixedInstruction(cmp)
			return
		}

	case AsmMov:
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		// is invalid mov instruction
		if isit && isitdst {
			mov1 := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			mov2 := AsmMov{
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			f.AppendFixedInstruction(mov1)
			f.AppendFixedInstruction(mov2)
			return
		}
	case AsmBinary:
		//check invalid check is both stack
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		if isit && isitdst && (ast.Op == Add || ast.Op == Sub) {
			mov := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			binary := AsmBinary{
				Op:  ast.Op,
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			// instructions[idx] = binary
			// instructions = append(instructions[:idx+1], append([]AsmInstruction{mov}, instructions[idx+1:]...)...)
			f.AppendFixedInstruction(mov)
			f.AppendFixedInstruction(binary)
			return
		}
		//FIX: this type is wrong
		if ast.Op == Mult && isitdst {
			mov := AsmMov{
				Src: dststack,
				Dst: Register{Reg: R11},
			}
			binary := AsmBinary{
				Op:  ast.Op,
				Src: ast.Src,
				Dst: Register{Reg: R11},
			}
			mov2 := AsmMov{
				Src: Register{Reg: R11},
				Dst: dststack,
			}

			f.AppendFixedInstruction(mov)
			f.AppendFixedInstruction(binary)
			f.AppendFixedInstruction(mov2)
			return
		}
	case Idiv:
		mov := AsmMov{
			Src: ast.Src,
			Dst: Register{Reg: R10},
		}
		idiv := Idiv{
			Src: Register{Reg: R10},
		}
		f.AppendFixedInstruction(mov)
		f.AppendFixedInstruction(idiv)
		return
	}
	f.AppendFixedInstruction(instr)
	return
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	allocStack := AllocateStack{
		Value: absInt(f.stackSize),
	}
	// fn.Irs = append([]AsmInstruction{allocStack}, fn.Irs...)
	f.fixedInstructions = append(f.fixedInstructions, allocStack)

	for i, instr := range fn.Irs {
		f.FixUpInInstruction(instr, f.fixedInstructions, i)
	}

	fn.Irs = f.fixedInstructions
	return fn
}

func (f *FixUpPassGen) FixUpProgram(program AsmProgram) AsmProgram {
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		asmFnDefs = append(asmFnDefs, f.FixUpInFn(fn))
	}
	return AsmProgram{AsmFnDef: asmFnDefs}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
