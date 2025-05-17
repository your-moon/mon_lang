package codegen

import (
	"github.com/your-moon/mn_compiler_go_version/util/utfconvert"
)

type TranslatePass struct{}

func NewTranslatePass() TranslatePass {
	return TranslatePass{}
}

func (f *TranslatePass) TranslateInInstr(instr AsmInstruction) AsmInstruction {
	switch ast := instr.(type) {
	case Label:
		ast.Ident = utfconvert.UtfConvert(ast.Ident)
		return ast
	case Call:
		ast.Ident = utfconvert.UtfConvert(ast.Ident)
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
	return AsmProgram{AsmFnDef: asmFnDefs, AsmExternFn: program.AsmExternFn}
}
