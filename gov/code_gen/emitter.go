package codegen

import "github.com/your-moon/mn_compiler_go_version/tackygen"

type MachineTarget string

type Emitter interface {
	Emit()
	starter()
	ending()
}

type AsmGen struct {
	Irs []AsmInstruction
}

func NewAsmGen() AsmGen {
	return AsmGen{}
}

func (a *AsmGen) EmitInstr(instr AsmInstruction) {
	a.Irs = append(a.Irs, instr)
}

func (a *AsmGen) GenAsm(program tackygen.TackyProgram) AsmProgram {
	asmprogram := AsmProgram{}

	asmfn := a.GenFn(program.FnDef)
	asmprogram.AsmFnDef = asmfn

	pass1 := NewReplacementPassGen()
	asmprogram = pass1.ReplacePseudosInProgram(asmprogram)

	pass2 := NewFixUpPassGen()
	asmprogram = pass2.FixUpProgram(asmprogram)

	return asmprogram
}

func (a *AsmGen) GenFn(fn tackygen.TackyFn) AsmFnDef {
	asmfn := AsmFnDef{}

	a.GenInstr(fn.Instructions)
	asmfn.Irs = a.Irs

	return asmfn
}

func (a *AsmGen) GenInstr(instrs []tackygen.Instruction) {

	for _, instr := range instrs {
		switch ast := instr.(type) {
		case tackygen.Return:
			mov := Mov{
				a.GenVal(ast.Value),
				Register{
					Reg: AX,
				},
			}
			ret := Return{}
			a.EmitInstr(mov)
			a.EmitInstr(ret)
		case tackygen.Unary:
			dst := a.GenVal(ast.Dst)
			mov := Mov{
				Src: a.GenVal(ast.Src),
				Dst: a.GenVal(ast.Dst),
			}
			unary := Unary{
				Op:  a.GenUnaryOp(ast.Op),
				Dst: dst,
			}
			a.EmitInstr(mov)
			a.EmitInstr(unary)
		}
	}

}

func (a *AsmGen) GenUnaryOp(op tackygen.UnaryOperator) AsmUnaryOperator {
	switch op {
	case tackygen.Complement:
		return Not
	case tackygen.Negate:
		return Neg
	default:
		panic("unimplemented tacky op on asm gen")
	}
}

func (a *AsmGen) GenVal(val tackygen.TackyVal) AsmOperand {
	switch ast := val.(type) {
	case tackygen.Constant:
		return Imm{Value: ast.Value}
	case tackygen.Var:
		return Pseudo{Ident: ast.Name}
	}

	return nil
}
