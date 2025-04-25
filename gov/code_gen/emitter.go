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
		case tackygen.Jump:
			jmp := Jmp{
				Ident: ast.Target,
			}
			a.EmitInstr(jmp)
		case tackygen.Label:
			lbl := Label{
				Ident: ast.Ident,
			}
			a.EmitInstr(lbl)
		case tackygen.Copy:
			mov := AsmMov{
				Src: a.GenASTVal(ast.Src),
				Dst: a.GenASTVal(ast.Dst),
			}
			a.EmitInstr(mov)
		case tackygen.JumpIfZero:
			cmp := Cmp{
				Src: Imm{Value: 0},
				Dst: a.GenASTVal(ast.Val),
			}
			jmpcc := JmpCC{
				CC:    E,
				Ident: ast.Ident,
			}
			a.EmitInstr(cmp)
			a.EmitInstr(jmpcc)
		case tackygen.JumpIfNotZero:
			cmp := Cmp{
				Src: Imm{Value: 0},
				Dst: a.GenASTVal(ast.Val),
			}
			jmpcc := JmpCC{
				CC:    NE,
				Ident: ast.Ident,
			}
			a.EmitInstr(cmp)
			a.EmitInstr(jmpcc)
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
			if ast.Op == tackygen.Not {
				cmp := Cmp{
					Src: Imm{Value: 0},
					Dst: a.GenASTVal(ast.Src),
				}
				mov := AsmMov{
					Src: Imm{Value: 0},
					Dst: a.GenASTVal(ast.Dst),
				}
				setcc := SetCC{
					CC: E,
					Op: a.GenASTVal(ast.Dst),
				}
				a.EmitInstr(cmp)
				a.EmitInstr(mov)
				a.EmitInstr(setcc)
				return
			}

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

func (a *AsmASTGen) ConvOpToCond(op tackygen.TackyBinaryOp) CondCode {
	switch op {
	case tackygen.GreaterThan:
		return G
	case tackygen.GreaterThanEqual:
		return GE
	case tackygen.LessThan:
		return L
	case tackygen.LessThanEqual:
		return LE
	case tackygen.Equal:
		return E
	case tackygen.NotEqual:
		return NE
	default:
		panic("the op is not relational op")

	}
}

func (a *AsmASTGen) GenASTBinary(instr tackygen.Binary) {
	//is relational op
	if instr.Op == tackygen.GreaterThan || instr.Op == tackygen.GreaterThanEqual || instr.Op == tackygen.LessThan || instr.Op == tackygen.LessThanEqual || instr.Op == tackygen.Equal || instr.Op == tackygen.NotEqual {
		cmp := Cmp{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Src2),
		}
		mov := AsmMov{
			Src: Imm{Value: 0},
			Dst: a.GenASTVal(instr.Dst),
		}
		setcc := SetCC{
			CC: a.ConvOpToCond(instr.Op),
			Op: a.GenASTVal(instr.Dst),
		}
		a.EmitInstr(cmp)
		a.EmitInstr(mov)
		a.EmitInstr(setcc)
		return
	}
	if instr.Op == tackygen.Remainder {
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: Register{Reg: AX},
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
		return
	} else if instr.Op == tackygen.Div {
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: Register{Reg: AX},
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
		return
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
		return
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
