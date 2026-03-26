package codegen

import (
	"github.com/your-moon/mon_lang/util/utfconvert"
)

type TranslatePass struct{}

func NewTranslatePass() TranslatePass {
	return TranslatePass{}
}

func (f *TranslatePass) translateOperand(op AsmOperand) AsmOperand {
	if rip, ok := op.(RipRelative); ok {
		rip.Label = utfconvert.UtfConvert(rip.Label)
		return rip
	}
	return op
}

func (f *TranslatePass) TranslateInInstr(instr AsmInstruction) AsmInstruction {
	switch ast := instr.(type) {
	case Label:
		ast.Ident = utfconvert.UtfConvert(ast.Ident)
		return ast
	case Call:
		ast.Ident = utfconvert.UtfConvert(ast.Ident)
		return ast
	case AsmMov:
		ast.Src = f.translateOperand(ast.Src)
		ast.Dst = f.translateOperand(ast.Dst)
		return ast
	case AsmBinary:
		ast.Src = f.translateOperand(ast.Src)
		ast.Dst = f.translateOperand(ast.Dst)
		return ast
	case Cmp:
		ast.Src = f.translateOperand(ast.Src)
		ast.Dst = f.translateOperand(ast.Dst)
		return ast
	case AsmMovSx:
		ast.Src = f.translateOperand(ast.Src)
		ast.Dst = f.translateOperand(ast.Dst)
		return ast
	case SetCC:
		ast.Op = f.translateOperand(ast.Op)
		return ast
	case Unary:
		ast.Dst = f.translateOperand(ast.Dst)
		return ast
	case Idiv:
		ast.Src = f.translateOperand(ast.Src)
		return ast
	case Push:
		ast.Op = f.translateOperand(ast.Op)
		return ast
	default:
		return instr
	}
}

func (f *TranslatePass) TranslateInFn(fn AsmFnDef) AsmFnDef {
	fn.Ident = utfconvert.UtfConvert(fn.Ident)

	for i, instr := range fn.Irs {
		fn.Irs[i] = f.TranslateInInstr(instr)
	}

	return fn
}

func (f *TranslatePass) TranslateProgram(program AsmProgram) AsmProgram {
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		asmFnDefs = append(asmFnDefs, f.TranslateInFn(fn))
	}
	for i, fn := range program.AsmExternFn {
		translated := utfconvert.UtfConvert(fn.Name)
		fn.Name = translated
		program.AsmExternFn[i] = fn
	}
	for i, gv := range program.GlobalVars {
		program.GlobalVars[i].Label = utfconvert.UtfConvert(gv.Label)
	}
	return AsmProgram{AsmFnDef: asmFnDefs, AsmExternFn: program.AsmExternFn, GlobalVars: program.GlobalVars}
}
