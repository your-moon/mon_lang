package codegen

type FixUpPassGen struct{}

func NewFixUpPassGen() FixUpPassGen {
	return FixUpPassGen{}
}

// in golang if you give array to param, its call by ref that means instructions is mutable
func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction, instructions []AsmInstruction, idx int) {
	mov, isit := instr.(Mov)

	if isit {
		srcstack, isit := mov.Src.(Stack)
		dststack, isitdst := mov.Dst.(Stack)
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
		}
	}
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	for i, instr := range fn.Irs {
		f.FixUpInInstruction(instr, fn.Irs, i)
	}
	return fn
}

func (f *FixUpPassGen) FixUpProgram(program AsmProgram) AsmProgram {
	program.AsmFnDef = f.FixUpInFn(program.AsmFnDef)
	return program
}
