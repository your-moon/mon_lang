package codegen

import "github.com/your-moon/mn_compiler_go_version/tackygen"

type MachineTarget string

type Emitter interface {
	Emit()
	starter()
	ending()
}

type AsmASTGen struct {
	Irs []AsmInstruction
}

func NewAsmGen() AsmASTGen {
	return AsmASTGen{}
}

func (a *AsmASTGen) EmitInstr(instr AsmInstruction) {
	a.Irs = append(a.Irs, instr)
}

func (a *AsmASTGen) GenASTAsm(program tackygen.TackyProgram) AsmProgram {
	asmprogram := AsmProgram{}

	asmfn := a.GenASTFn(program.FnDef)
	asmprogram.AsmFnDef = asmfn

	pass1 := NewReplacementPassGen()
	asmprogram = pass1.ReplacePseudosInProgram(asmprogram)

	pass2 := NewFixUpPassGen(pass1.CurrentOffset)
	asmprogram = pass2.FixUpProgram(asmprogram)

	return asmprogram
}

func (a *AsmASTGen) GenASTFn(fn tackygen.TackyFn) AsmFnDef {
	asmfn := AsmFnDef{}

	a.GenASTInstr(fn.Instructions)
	asmfn.Irs = a.Irs
	asmfn.Ident = fn.Name

	return asmfn
}

func (a *AsmASTGen) GenASTInstr(instrs []tackygen.Instruction) {

	for _, instr := range instrs {
		switch ast := instr.(type) {
		case tackygen.Return:
			mov := Mov{
				a.GenASTVal(ast.Value),
				Register{
					Reg: AX,
				},
			}
			ret := Return{}
			a.EmitInstr(mov)
			a.EmitInstr(ret)
		case tackygen.Binary:
			a.GenASTBinary(ast)
		case tackygen.Unary:
			dst := a.GenASTVal(ast.Dst)
			mov := Mov{
				Src: a.GenASTVal(ast.Src),
				Dst: a.GenASTVal(ast.Dst),
			}
			unary := Unary{
				Op:  a.GenASTUnaryOp(ast.Op),
				Dst: dst,
			}
			a.EmitInstr(mov)
			a.EmitInstr(unary)
		}
	}

}
func (a *AsmASTGen) GenASTBinary(instr tackygen.Binary) {
	if instr.Op == tackygen.Remainder {
		mov := Mov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		cdq := Cdq{}
		idiv := Idiv{
			Src: a.GenASTVal(instr.Src2),
		}
		mov2 := Mov{
			Src: Register{
				Reg: DX,
			},

			Dst: a.GenASTVal(instr.Dst),
		}
		a.EmitInstr(mov)
		a.EmitInstr(cdq)
		a.EmitInstr(idiv)
		a.EmitInstr(mov2)

	} else if instr.Op == tackygen.Div {
		mov := Mov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		cdq := Cdq{}
		idiv := Idiv{
			Src: a.GenASTVal(instr.Src2),
		}
		mov2 := Mov{
			Src: Register{
				Reg: AX,
			},

			Dst: a.GenASTVal(instr.Dst),
		}
		a.EmitInstr(mov)
		a.EmitInstr(cdq)
		a.EmitInstr(idiv)
		a.EmitInstr(mov2)
	} else {
		mov := Mov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		binary := Binary{
			Op:  BinaryOp(instr.Op),
			Src: a.GenASTVal(instr.Src2),
			Dst: a.GenASTVal(instr.Dst),
		}
		a.EmitInstr(mov)
		a.EmitInstr(binary)

	}
}

func (a *AsmASTGen) GenASTUnaryOp(op tackygen.UnaryOperator) AsmUnaryOperator {
	switch op {
	case tackygen.Complement:
		return Not
	case tackygen.Negate:
		return Neg
	default:
		panic("unimplemented tacky op on asm gen")
	}
}

func (a *AsmASTGen) GenASTVal(val tackygen.TackyVal) AsmOperand {
	switch ast := val.(type) {
	case tackygen.Constant:
		return Imm{Value: ast.Value}
	case tackygen.Var:
		return Pseudo{Ident: ast.Name}
	}

	return nil
}
