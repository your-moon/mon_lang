package codegen

type FixUpPassGen struct{}

func NewFixUpPassGen() FixUpPassGen {
	return FixUpPassGen{}
}

func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction, fn AsmFnDef, idx int) {
	mov, isit := instr.(Mov)

	if isit {
		srcstack, isit := mov.Src.(Stack)
		dststack, isitdst := mov.Dst.(Stack)
		if isit && isitdst {
			mov1 := Mov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			mov2 := Mov{
				Src: Register{Reg: R10},
				Dst: dststack,
			}

			//FIX: FIX
			// Split the instructions at idx and insert mov1 and mov2
			right := fn.Irs[idx:]
			fn.Irs = fn.Irs[:idx]
			fn.Irs = append(fn.Irs, mov1, mov2)
			fn.Irs = append(fn.Irs, right...)
		}
	}
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	for i, instr := range fn.Irs {
		f.FixUpInInstruction(instr, fn, i)

	}
	return fn
}
func (f *FixUpPassGen) FixUpProgram(program AsmProgram) AsmProgram {
	program.AsmFnDef = f.FixUpInFn(program.AsmFnDef)
	return program
}
