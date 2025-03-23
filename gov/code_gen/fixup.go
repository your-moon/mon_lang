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
			return instructions
		}
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
