package codegen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/base"
	"github.com/your-moon/mn_compiler_go_version/code_gen/asmsymbol"
	"github.com/your-moon/mn_compiler_go_version/code_gen/asmtype"
	"github.com/your-moon/mn_compiler_go_version/mconstant"
	"github.com/your-moon/mn_compiler_go_version/mtypes"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/tackygen"
)

type MachineTarget string

type Emitter interface {
	Emit()
	starter()
	ending()
}

type AsmASTGen struct {
	Registers   []AsmRegister
	SymbolTable *symbols.SymbolTable
}

func NewAsmGen(table *symbols.SymbolTable) AsmASTGen {
	return AsmASTGen{
		Registers:   []AsmRegister{DI, SI, DX, CX, R8, R9},
		SymbolTable: table,
	}
}

func (a *AsmASTGen) GenASTAsm(program tackygen.TackyProgram, symbolTable *symbols.SymbolTable, asmSymbols *asmsymbol.SymbolTable) AsmProgram {
	asmprogram := AsmProgram{}

	for _, fn := range program.ExternDefs {
		asmfn := a.GenASTExternFn(fn)
		asmprogram.AsmExternFn = append(asmprogram.AsmExternFn, asmfn)
	}

	for _, fn := range program.FnDefs {
		asmfn := a.GenASTFn(fn)
		asmprogram.AsmFnDef = append(asmprogram.AsmFnDef, asmfn)
	}

	for i, v := range symbolTable.Entries {
		switch v.Type.(type) {
		case *mtypes.FnType:
			asmSymbols.AddFun(i, true)
		default:
			convType := a.ConvType(v.Type)
			asmSymbols.AddVar(i, convType, false)
		}
	}
	// END ON INIT PASS //

	if base.Debug {
		fmt.Println("---- ASMAST ----:")
		for _, fn := range asmprogram.AsmExternFn {
			fmt.Println(fn.Ir())
		}
		for _, fn := range asmprogram.AsmFnDef {
			for _, instr := range fn.Irs {
				fmt.Println(instr.Ir())
			}
		}
	}

	pass1 := NewReplacementPassGen()
	asmprogram = pass1.ReplacePseudosInProgram(asmprogram, symbolTable)

	if base.Debug {
		fmt.Println("---- ASMAST AFTER PSEUDO REPLACEMENT ----:")
		for _, fn := range asmprogram.AsmFnDef {
			for _, instr := range fn.Irs {
				fmt.Println(instr.Ir())
			}
		}
	}

	pass2 := NewFixUpPassGen(symbolTable)
	asmprogram = pass2.FixUpProgram(asmprogram)

	pass3 := NewTranslatePass()
	asmprogram = pass3.TranslateProgram(asmprogram)

	if base.Debug {
		fmt.Println("---- ASMAST AFTER FIXUP ----:")
		for _, fn := range asmprogram.AsmFnDef {
			for _, instr := range fn.Irs {
				fmt.Println(instr.Ir())
			}
		}
	}

	return asmprogram
}

func (a *AsmASTGen) GenASTExternFn(fn tackygen.TackyFn) AsmExternFn {
	asmfn := AsmExternFn{
		Name: fn.Name,
	}
	return asmfn
}

func (a *AsmASTGen) passInStack(param tackygen.TackyVal) []AsmInstruction {
	switch valtype := param.(type) {
	case tackygen.Var:
		asmArg := a.GenASTVal(valtype)
		switch asmArg.(type) {
		case Register:
		case Imm:
			push := Push{
				Op: asmArg,
			}
			return []AsmInstruction{push}
		default:
			pType := a.AsmType(param)
			mov := AsmMov{
				Type: pType,
				Src:  asmArg,
				Dst:  Register{Reg: AX},
			}
			push := Push{
				Op: Register{Reg: AX},
			}
			return []AsmInstruction{mov, push}
		}
	}
	return []AsmInstruction{}
}

func (a *AsmASTGen) AsmType(val tackygen.TackyVal) asmtype.AsmType {
	switch valtype := val.(type) {
	case tackygen.Constant:
		switch valtype.Value.(type) {
		case mconstant.Int64:
			return &asmtype.QuadWord{}
		case mconstant.Int32:
			return &asmtype.LongWord{}
		default:
			panic("unimplemented")
		}
	case *tackygen.Var:
		vtype := a.SymbolTable.Get(valtype.Name)
		valType := a.ConvType(vtype.Type)
		return valType
	default:
		panic("unimplemented")
	}
}

func (a *AsmASTGen) passInRegisters(paramIdx int, param tackygen.TackyVal) []AsmInstruction {
	passRegister := ""
	for _, register := range a.Registers {
		if register.String() == a.Registers[paramIdx].String() {
			passRegister = register.String()
			break
		}
	}

	if passRegister == "" {
		return []AsmInstruction{}
	}

	pType := a.AsmType(param)
	mov := AsmMov{
		Type: pType,
		Src:  Register{Reg: AsmRegister(passRegister)},
		Dst:  a.GenASTVal(param),
	}
	return []AsmInstruction{mov}
}

func (a *AsmASTGen) passParams(fn tackygen.TackyFn) (tackygen.TackyFn, []AsmInstruction) {
	ir := []AsmInstruction{}
	registerParams := []tackygen.TackyVal{}
	stackParams := []tackygen.TackyVal{}

	for _, param := range fn.Params {
		if len(registerParams) < 6 {
			registerParams = append(registerParams, param)
		} else {
			stackParams = append(stackParams, param)
		}
	}

	for i, param := range registerParams {
		ir = append(ir, a.passInRegisters(i, param)...)
	}

	for _, param := range stackParams {
		ir = append(ir, a.passInStack(param)...)
	}

	return fn, ir
}

func (a *AsmASTGen) splitArgs(args []tackygen.TackyVal) ([]tackygen.TackyVal, []tackygen.TackyVal) {
	registerArgs := []tackygen.TackyVal{}
	stackArgs := []tackygen.TackyVal{}

	for _, arg := range args {
		if len(registerArgs) < 6 {
			registerArgs = append(registerArgs, arg)
		} else {
			stackArgs = append(stackArgs, arg)
		}
	}

	return registerArgs, stackArgs
}

func (a *AsmASTGen) convertFnCall(fn tackygen.FnCall) []AsmInstruction {
	irs := []AsmInstruction{}
	argRegisters := []AsmRegister{DI, SI, DX, CX, R8, R9}
	stackPadding := 0

	registerArgs, stackArgs := a.splitArgs(fn.Args)

	if len(stackArgs)%2 != 0 {
		stackPadding = 8
	}

	for i, arg := range registerArgs {
		r := argRegisters[i]
		mov := AsmMov{
			Src: a.GenASTVal(arg),
			Dst: Register{Reg: r},
		}
		irs = append(irs, mov)
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
			irs = append(irs, push)
		default:
			mov := AsmMov{
				Src: asmArg,
				Dst: Register{Reg: AX},
			}
			irs = append(irs, mov)
			push := Push{
				Op: Register{Reg: AX},
			}
			irs = append(irs, push)
		}
	}

	call := Call{
		Ident: fn.Name,
	}
	irs = append(irs, call)

	bytesToRemove := 8*len(stackArgs) + stackPadding
	if bytesToRemove != 0 {
		deallocate := DeallocateStack{
			Value: bytesToRemove,
		}
		irs = append(irs, deallocate)
	}

	asmDst := a.GenASTVal(fn.Dst)
	mov := AsmMov{
		Src: Register{Reg: AX},
		Dst: asmDst,
	}
	irs = append(irs, mov)

	return irs
}

func (a *AsmASTGen) GenASTFn(fn tackygen.TackyFn) AsmFnDef {
	asmfn := AsmFnDef{}

	fn, paramIrs := a.passParams(fn)
	asmfn.Irs = append(asmfn.Irs, paramIrs...)

	for _, instr := range fn.Instructions {
		asmfn.Irs = append(asmfn.Irs, a.GenASTInstr(instr)...)
	}
	asmfn.Ident = fn.Name
	return asmfn
}

func (a *AsmASTGen) GenASTInstr(instr tackygen.Instruction) []AsmInstruction {
	switch ast := instr.(type) {
	case tackygen.FnCall:
		return a.convertFnCall(ast)
	case tackygen.Jump:
		jmp := Jmp{
			Ident: ast.Target,
		}
		return []AsmInstruction{jmp}
	case tackygen.Label:
		lbl := Label{
			Ident: ast.Ident,
		}
		return []AsmInstruction{lbl}
	case tackygen.Copy:
		Type := a.AsmType(ast.Src)
		mov := AsmMov{
			Type: Type,
			Src:  a.GenASTVal(ast.Src),
			Dst:  a.GenASTVal(ast.Dst),
		}
		return []AsmInstruction{mov}
	case tackygen.JumpIfZero:
		Type := a.AsmType(ast.Val)
		cmp := Cmp{
			Type: Type,
			Src:  Imm{Value: 0},
			Dst:  a.GenASTVal(ast.Val),
		}
		jmpcc := JmpCC{
			CC:    E,
			Ident: ast.Ident,
		}
		return []AsmInstruction{cmp, jmpcc}
	case tackygen.JumpIfNotZero:
		Type := a.AsmType(ast.Val)
		cmp := Cmp{
			Type: Type,
			Src:  Imm{Value: 0},
			Dst:  a.GenASTVal(ast.Val),
		}
		jmpcc := JmpCC{
			CC:    NE,
			Ident: ast.Ident,
		}
		return []AsmInstruction{cmp, jmpcc}
	case tackygen.Return:
		Type := a.AsmType(ast.Value)
		mov := AsmMov{
			Type: Type,
			Src:  a.GenASTVal(ast.Value),
			Dst: Register{
				Reg: AX,
			},
		}
		ret := Return{}
		return []AsmInstruction{mov, ret}
	case tackygen.Binary:
		return a.GenASTBinary(ast)
	case tackygen.Unary:
		SrcT := a.AsmType(ast.Src)
		DstT := a.AsmType(ast.Dst)
		if ast.Op == tackygen.Not {
			cmp := Cmp{
				Type: SrcT,
				Src:  Imm{Value: 0},
				Dst:  a.GenASTVal(ast.Src),
			}
			mov := AsmMov{
				Type: DstT,
				Src:  Imm{Value: 0},
				Dst:  a.GenASTVal(ast.Dst),
			}
			setcc := SetCC{
				CC: E,
				Op: a.GenASTVal(ast.Dst),
			}
			return []AsmInstruction{cmp, mov, setcc}
		}

		dst := a.GenASTVal(ast.Dst)
		mov := AsmMov{
			Type: SrcT,
			Src:  a.GenASTVal(ast.Src),
			Dst:  a.GenASTVal(ast.Dst),
		}
		unary := Unary{
			Type: SrcT,
			Op:   a.GenASTUnaryOp(ast.Op),
			Dst:  dst,
		}
		return []AsmInstruction{mov, unary}
	default:
		panic("unimplemented tacky instruction on asm gen")
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

func (a *AsmASTGen) GenASTBinary(instr tackygen.Binary) []AsmInstruction {
	Src1T := a.AsmType(instr.Src1)
	DstT := a.AsmType(instr.Dst)
	//is relational op
	if instr.Op == tackygen.GreaterThan || instr.Op == tackygen.GreaterThanEqual || instr.Op == tackygen.LessThan || instr.Op == tackygen.LessThanEqual || instr.Op == tackygen.Equal || instr.Op == tackygen.NotEqual {

		cmp := Cmp{
			Type: Src1T,
			Src:  a.GenASTVal(instr.Src2),
			Dst:  a.GenASTVal(instr.Src1),
		}
		mov := AsmMov{
			Type: DstT,
			Src:  Imm{Value: 0},
			Dst:  a.GenASTVal(instr.Dst),
		}
		setcc := SetCC{
			CC: a.ConvOpToCond(instr.Op),
			Op: a.GenASTVal(instr.Dst),
		}
		return []AsmInstruction{cmp, mov, setcc}
	}
	if instr.Op == tackygen.Modulo {

		mov := AsmMov{
			Type: Src1T,
			Src:  a.GenASTVal(instr.Src1),
			Dst:  Register{Reg: AX},
		}
		cdq := Cdq{
			Type: Src1T,
		}
		idiv := Idiv{
			Type: Src1T,
			Src:  a.GenASTVal(instr.Src2),
		}
		mov2 := AsmMov{
			Type: Src1T,
			Src: Register{
				Reg: DX,
			},

			Dst: a.GenASTVal(instr.Dst),
		}
		return []AsmInstruction{mov, cdq, idiv, mov2}
	} else if instr.Op == tackygen.Div {
		mov := AsmMov{
			Type: Src1T,
			Src:  a.GenASTVal(instr.Src1),
			Dst:  Register{Reg: AX},
		}
		cdq := Cdq{

			Type: Src1T,
		}
		idiv := Idiv{
			Type: Src1T,
			Src:  a.GenASTVal(instr.Src2),
		}
		mov2 := AsmMov{
			Type: Src1T,
			Src: Register{
				Reg: AX,
			},

			Dst: a.GenASTVal(instr.Dst),
		}
		return []AsmInstruction{mov, cdq, idiv, mov2}
	} else {
		mov := AsmMov{
			Src: a.GenASTVal(instr.Src1),
			Dst: a.GenASTVal(instr.Dst),
		}
		binary := AsmBinary{
			Type: Src1T,
			Op:   a.GenASTBinaryOp(instr.Op),
			Src:  a.GenASTVal(instr.Src2),
			Dst:  a.GenASTVal(instr.Dst),
		}
		return []AsmInstruction{mov, binary}
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
		return Imm{Value: ast.Value.GetValue()}
	case tackygen.Var:
		return Pseudo{Ident: ast.Name}
	}

	return nil
}

func (a *AsmASTGen) ConvType(val mtypes.Type) asmtype.AsmType {
	switch val.(type) {
	case *mtypes.IntType:
		return &asmtype.LongWord{}
	case *mtypes.LongType:
		return &asmtype.QuadWord{}
	case *mtypes.FnType:
		panic("fn type should not be here")
	}

	return nil
}
