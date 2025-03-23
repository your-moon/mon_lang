package codegen

type FixUpPassGen struct {
	stackSize int
}

func NewFixUpPassGen(stackSize int) FixUpPassGen {
	return FixUpPassGen{
		stackSize: stackSize,
	}
}

// in golang if you give array to param, its call by ref that means instructions is mutable
func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction, instructions []AsmInstruction, idx int) []AsmInstruction {
	switch ast := instr.(type) {
	case Mov:
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		// is invalid mov instruction
		if isit && isitdst {
			mov1 := Mov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			mov2 := Mov{
				Src: Register{Reg: R10},
				Dst: dststack,
			}

			instructions[idx] = mov1
			instructions = append(instructions[:idx+1], append([]AsmInstruction{mov2}, instructions[idx+1:]...)...)
			return instructions
		}
	case Binary:
		//check invalid check is both stack
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		if isit && isitdst && (ast.Op == Add || ast.Op == Sub) {
			mov := Mov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			binary := Binary{
				Op:  ast.Op,
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			instructions[idx] = mov
			instructions = append(instructions[:idx+1], append([]AsmInstruction{binary}, instructions[idx+1:]...)...)
		}
		//FIX: this type is wrong
		if ast.Op == Mult {
			mov := Mov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			binary := Binary{
				Op:  ast.Op,
				Src: ast.Src,
				Dst: Register{Reg: R10},
			}
			mov2 := Mov{
				Src: Register{Reg: R10},
				Dst: ast.Dst,
			}
			instructions[idx] = mov
			instructions = append(instructions[:idx+1], append([]AsmInstruction{binary}, instructions[idx+1:]...)...)
			instructions = append(instructions[:idx+2], append([]AsmInstruction{mov2}, instructions[idx+1:]...)...)
		}
	case Idiv:
		mov := Mov{
			Src: ast.Src,
			Dst: Register{Reg: R10},
		}
		idiv := Idiv{
			Src: Register{Reg: R10},
		}
		instructions[idx] = mov
		instructions = append(instructions[:idx+1], append([]AsmInstruction{idiv}, instructions[idx+1:]...)...)
	}

	return instructions
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	allocStack := AllocateStack{
		Value: absInt(f.stackSize),
	}
	fn.Irs = append([]AsmInstruction{allocStack}, fn.Irs...)
	for i, instr := range fn.Irs {
		fn.Irs = f.FixUpInInstruction(instr, fn.Irs, i)
	}
	return fn
}

func (f *FixUpPassGen) FixUpProgram(program AsmProgram) AsmProgram {
	program.AsmFnDef = f.FixUpInFn(program.AsmFnDef)
	return program
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
