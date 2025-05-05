package codegen

import (
	semanticanalysis "github.com/your-moon/mn_compiler_go_version/semantic_analysis"
	"github.com/your-moon/mn_compiler_go_version/util/roundingutil"
)

type FixUpPassGen struct {
	symbolTable *semanticanalysis.SymbolTable
}

func NewFixUpPassGen(symbolTable *semanticanalysis.SymbolTable) FixUpPassGen {
	return FixUpPassGen{
		symbolTable: symbolTable,
	}
}

// in golang if you give array to param, its call by ref that means instructions is mutable
func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction) []AsmInstruction {
	switch ast := instr.(type) {
	case Cmp:
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		// is invalid mov instruction
		if isit && isitdst {
			mov1 := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			cmp := Cmp{
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			return []AsmInstruction{mov1, cmp}
		}

		dstimm, isitdst := ast.Dst.(Imm)
		if isitdst {
			mov1 := AsmMov{
				Src: dstimm,
				Dst: Register{Reg: R11},
			}
			cmp := Cmp{
				Src: ast.Src,
				Dst: Register{Reg: R11},
			}
			return []AsmInstruction{mov1, cmp}
		}
		return []AsmInstruction{ast}
	case AsmMov:
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		// is invalid mov instruction
		if isit && isitdst {
			mov1 := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			mov2 := AsmMov{
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			return []AsmInstruction{mov1, mov2}
		}
		return []AsmInstruction{ast}
	case AsmBinary:
		//check invalid check is both stack
		srcstack, isit := ast.Src.(Stack)
		dststack, isitdst := ast.Dst.(Stack)
		if isit && isitdst && (ast.Op == Add || ast.Op == Sub) {
			mov := AsmMov{
				Src: srcstack,
				Dst: Register{Reg: R10},
			}
			binary := AsmBinary{
				Op:  ast.Op,
				Src: Register{Reg: R10},
				Dst: dststack,
			}
			return []AsmInstruction{mov, binary}
		}
		//FIX: this type is wrong
		if ast.Op == Mult && isitdst {
			mov := AsmMov{
				Src: dststack,
				Dst: Register{Reg: R11},
			}
			binary := AsmBinary{
				Op:  ast.Op,
				Src: ast.Src,
				Dst: Register{Reg: R11},
			}
			mov2 := AsmMov{
				Src: Register{Reg: R11},
				Dst: dststack,
			}

			return []AsmInstruction{mov, binary, mov2}
		}
		return []AsmInstruction{ast}
	case Idiv:
		mov := AsmMov{
			Src: ast.Src,
			Dst: Register{Reg: R10},
		}
		idiv := Idiv{
			Src: Register{Reg: R10},
		}
		return []AsmInstruction{mov, idiv}
	default:
		return []AsmInstruction{instr}
	}
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	allocStack := AllocateStack{
		Value: roundingutil.RoundAwayFromZero(16, f.symbolTable.Get(fn.Ident).StackFrameSize),
	}

	fixedInstructions := []AsmInstruction{}
	fixedInstructions = append(fixedInstructions, allocStack)

	for _, instr := range fn.Irs {
		fixedInstructions = append(fixedInstructions, f.FixUpInInstruction(instr)...)
	}

	fn.Irs = fixedInstructions

	return fn
}

func (f *FixUpPassGen) FixUpProgram(program AsmProgram) AsmProgram {
	asmFnDefs := []AsmFnDef{}
	for _, fn := range program.AsmFnDef {
		asmFnDefs = append(asmFnDefs, f.FixUpInFn(fn))
	}
	return AsmProgram{AsmFnDef: asmFnDefs}
}
