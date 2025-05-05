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
	Irs       []AsmInstruction
	Registers []AsmRegister
}

func NewAsmGen() AsmASTGen {
	return AsmASTGen{
		Irs:       []AsmInstruction{},
		Registers: []AsmRegister{DI, SI, DX, CX, R8, R9},
	}
}

func (a *AsmASTGen) EmitInstr(instr AsmInstruction) {
	a.Irs = append(a.Irs, instr)
}

func (a *AsmASTGen) GenASTAsm(program tackygen.TackyProgram) AsmProgram {
	asmprogram := AsmProgram{}

	for _, fn := range program.FnDefs {
		asmfn := a.GenASTFn(fn)
		asmprogram.AsmFnDef = append(asmprogram.AsmFnDef, asmfn)
	}

	fmt.Println("---- ASMAST ----:")
	for _, fn := range asmprogram.AsmFnDef {
		for _, instr := range fn.Irs {
			fmt.Println(instr.Ir())
		}
	}

	pass1 := NewReplacementPassGen()
	asmprogram = pass1.ReplacePseudosInProgram(asmprogram)

	fmt.Println("---- ASMAST AFTER PSEUDO REPLACEMENT ----:")
	for _, fn := range asmprogram.AsmFnDef {
		for _, instr := range fn.Irs {
			fmt.Println(instr.Ir())
		}
	}

	pass2 := NewFixUpPassGen(pass1.CurrentOffset)
	asmprogram = pass2.FixUpProgram(asmprogram)

	fmt.Println("---- ASMAST AFTER FIXUP ----:")
	for _, fn := range asmprogram.AsmFnDef {
		for _, instr := range fn.Irs {
			fmt.Println(instr.Ir())
		}
	}

	return asmprogram
}

func (a *AsmASTGen) passInStack(param tackygen.TackyVal) tackygen.TackyVal {
	switch valtype := param.(type) {
	case tackygen.Var:
		asmArg := a.GenASTVal(valtype)
		switch asmArg.(type) {
		case Register:
		case Imm:
			push := Push{
				Op: asmArg,
			}
			a.EmitInstr(push)
		default:
			mov := AsmMov{
				Src: asmArg,
				Dst: Register{Reg: AX},
			}
			a.EmitInstr(mov)
		}
	}
	return param
}

func (a *AsmASTGen) passInRegisters(param tackygen.TackyVal) tackygen.TackyVal {
	switch valtype := param.(type) {
	case tackygen.Var:
		passRegister := ""
		for _, register := range a.Registers {
			if register.String() == valtype.Name {
				passRegister = register.String()
			}
		}
		mov := AsmMov{
			Src: Imm{Value: 0},
			Dst: Register{Reg: AsmRegister(passRegister)},
		}
		a.EmitInstr(mov)
	}
	return param
}

func (a *AsmASTGen) passParams(fn tackygen.TackyFn) tackygen.TackyFn {
	registerParams := []tackygen.TackyVal{}
	stackParams := []tackygen.TackyVal{}

	for _, param := range fn.Params {
		if len(registerParams) < 6 {
			registerParams = append(registerParams, param)
		} else {
			stackParams = append(stackParams, param)
		}
	}

	for _, param := range registerParams {
		a.passInRegisters(param)
	}

	for _, param := range stackParams {
		a.passInStack(param)
	}

	return fn
}

func (a *AsmASTGen) GenASTFn(fn tackygen.TackyFn) AsmFnDef {
	asmfn := AsmFnDef{}
	fn = a.passParams(fn)
	a.GenASTInstr(fn.Instructions)
	asmfn.Irs = a.Irs
	asmfn.Ident = fn.Name

	return asmfn
}

func (a *AsmASTGen) GenASTInstr(instrs []tackygen.Instruction) {

	for _, instr := range instrs {
		switch ast := instr.(type) {
		case tackygen.FnCall:
			argRegisters := []AsmRegister{DI, SI, DX, CX, R8, R9}
			stackPadding := 0

			registerArgs := []tackygen.TackyVal{}
			stackArgs := []tackygen.TackyVal{}
			if len(ast.Args) <= 6 {
				registerArgs = ast.Args
				stackArgs = []tackygen.TackyVal{}
			} else {
				registerArgs = ast.Args[:6]
				stackArgs = ast.Args[6:]
			}

			if len(stackArgs)%2 != 0 {
				stackPadding = 8
			} else {
				stackPadding = 0
			}

			if stackPadding != 0 {
				a.EmitInstr(AllocateStack{Value: stackPadding})
			}

			for i, arg := range registerArgs {
				r := argRegisters[i]
				mov := AsmMov{
					Src: a.GenASTVal(arg),
					Dst: Register{Reg: r},
				}
				a.EmitInstr(mov)
			}

			reversedArgs := make([]tackygen.TackyVal, len(stackArgs))
			for i, arg := range stackArgs {
				reversedArgs[len(stackArgs)-1-i] = arg
			}

			for _, arg := range reversedArgs {
				asmArg := a.GenASTVal(arg)
				switch asmArg.(type) {
				case Register:
				case Imm:
					push := Push{
						Op: asmArg,
					}
					a.EmitInstr(push)
				default:
					mov := AsmMov{
						Src: asmArg,
						Dst: Register{Reg: AX},
					}
					a.EmitInstr(mov)
					push := Push{
						Op: Register{Reg: AX},
					}
					a.EmitInstr(push)
				}
			}

			call := Call{
				Ident: ast.Name,
			}
			a.EmitInstr(call)

			bytesToRemove := 8*len(stackArgs) + stackPadding
			if bytesToRemove != 0 {
				a.EmitInstr(DeallocateStack{Value: bytesToRemove})
			}
			asmDst := a.GenASTVal(ast.Dst)
			mov := AsmMov{
				Src: asmDst,
				Dst: Register{Reg: AX},
			}
			a.EmitInstr(mov)
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
			Src: a.GenASTVal(instr.Src2),
			Dst: a.GenASTVal(instr.Src1),
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
