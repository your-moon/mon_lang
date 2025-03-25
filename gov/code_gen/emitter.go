package codegen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/tackygen"
)

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

	fmt.Println("---- ASMAST ----:")
	for _, instr := range asmprogram.AsmFnDef.Irs {
		fmt.Println(instr.Ir())
	}

	pass1 := NewReplacementPassGen()
	asmprogram = pass1.ReplacePseudosInProgram(asmprogram)

	fmt.Println("---- ASMAST AFTER PSEUDO REPLACEMENT ----:")
	for _, instr := range asmprogram.AsmFnDef.Irs {
		fmt.Println(instr.Ir())
	}

	pass2 := NewFixUpPassGen(pass1.CurrentOffset)
	asmprogram = pass2.FixUpProgram(asmprogram)

	fmt.Println("---- ASMAST AFTER FIXUP ----:")
	for _, instr := range asmprogram.AsmFnDef.Irs {
		fmt.Println(instr.Ir())
	}

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
			mov := AsmMov{
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
			mov := AsmMov{
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
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		cdq := Cdq{}
		idiv := Idiv{
			Src: a.GenASTVal(instr.Src2),
		}
		mov2 := AsmMov{
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
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		cdq := Cdq{}
		idiv := Idiv{
			Src: a.GenASTVal(instr.Src2),
		}
		mov2 := AsmMov{
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
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		binary := AsmBinary{
			Op:  a.GenASTBinaryOp(instr.Op),
			Src: a.GenASTVal(instr.Src2),
			Dst: a.GenASTVal(instr.Dst),
		}
		a.EmitInstr(mov)
		a.EmitInstr(binary)

	}
}

func (a *AsmASTGen) GenASTBinaryOp(op tackygen.TackyBinaryOp) AsmAstBinaryOp {
	switch op {
	case tackygen.Mul:
		return Mult
	case tackygen.Add:
		return Add
	case tackygen.Sub:
		return Sub
	}
	panic("unimplemented tacky op on asm gen")
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
