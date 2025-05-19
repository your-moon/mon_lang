package codegen

import (
	"github.com/your-moon/mn_compiler_go_version/code_gen/asmtype"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/util/roundingutil"
)

type FixUpPassGen struct {
	symbolTable *symbols.SymbolTable
}

func NewFixUpPassGen(symbolTable *symbols.SymbolTable) FixUpPassGen {
	return FixUpPassGen{
		symbolTable: symbolTable,
	}
}

// isLarge checks if an immediate value is too large for direct use
func isLarge(i int64) bool {
	return i > 0x7fffffff || i < -0x80000000
}

func (f *FixUpPassGen) FixUpInInstruction(instr AsmInstruction) []AsmInstruction {
	switch ast := instr.(type) {
	case AsmMov:
		// Handle quadword immediate to register
		if _, isQuadWord := ast.Type.(*asmtype.QuadWord); isQuadWord {
			if imm, isImm := ast.Src.(Imm); isImm {
				if isLarge(imm.Value) {
					return []AsmInstruction{
						AsmMov{
							Type: &asmtype.QuadWord{},
							Src:  imm,
							Dst:  Register{Reg: R10},
						},
						AsmMov{
							Type: &asmtype.QuadWord{},
							Src:  Register{Reg: R10},
							Dst:  ast.Dst,
						},
					}
				}
			}
		}

		// Handle longword immediate that's too large
		if _, isLongWord := ast.Type.(*asmtype.LongWord); isLongWord {
			if imm, isImm := ast.Src.(Imm); isImm {
				if isLarge(imm.Value) {
					// Reduce modulo 2^32 by zeroing out upper 32 bits
					reduced := imm.Value & 0xffffffff
					return []AsmInstruction{
						AsmMov{
							Type: &asmtype.LongWord{},
							Src:  Imm{Value: reduced},
							Dst:  ast.Dst,
						},
					}
				}
			}
		}

		// Handle stack to stack move
		if srcStack, isSrcStack := ast.Src.(Stack); isSrcStack {
			if dstStack, isDstStack := ast.Dst.(Stack); isDstStack {
				return []AsmInstruction{
					AsmMov{
						Type: ast.Type,
						Src:  srcStack,
						Dst:  Register{Reg: R10},
					},
					AsmMov{
						Type: ast.Type,
						Src:  Register{Reg: R10},
						Dst:  dstStack,
					},
				}
			}
		}
		return []AsmInstruction{ast}

	case AsmMovSx:
		// Handle immediate source with stack/data destination
		if imm, isImm := ast.Src.(Imm); isImm {
			if _, isStack := ast.Dst.(Stack); isStack {
				return []AsmInstruction{
					AsmMov{
						Type: &asmtype.LongWord{},
						Src:  imm,
						Dst:  Register{Reg: R10},
					},
					AsmMovSx{
						Src: Register{Reg: R10},
						Dst: Register{Reg: R11},
					},
					AsmMov{
						Type: &asmtype.QuadWord{},
						Src:  Register{Reg: R11},
						Dst:  ast.Dst,
					},
				}
			}
			return []AsmInstruction{
				AsmMov{
					Type: &asmtype.LongWord{},
					Src:  imm,
					Dst:  Register{Reg: R10},
				},
				AsmMovSx{
					Src: Register{Reg: R10},
					Dst: ast.Dst,
				},
			}
		}

		// Handle stack/data destination
		if _, isStack := ast.Dst.(Stack); isStack {
			return []AsmInstruction{
				AsmMovSx{
					Src: ast.Src,
					Dst: Register{Reg: R11},
				},
				AsmMov{
					Type: &asmtype.QuadWord{},
					Src:  Register{Reg: R11},
					Dst:  ast.Dst,
				},
			}
		}
		return []AsmInstruction{ast}

	case Idiv:
		// Handle immediate source
		if imm, isImm := ast.Src.(Imm); isImm {
			return []AsmInstruction{
				AsmMov{
					Type: ast.Type,
					Src:  imm,
					Dst:  Register{Reg: R10},
				},
				Idiv{
					Type: ast.Type,
					Src:  Register{Reg: R10},
				},
			}
		}
		return []AsmInstruction{ast}

	case AsmBinary:
		// Handle large immediate source for Add/Sub
		if (ast.Op == Add || ast.Op == Sub) && ast.Type.(*asmtype.QuadWord) != nil {
			if imm, isImm := ast.Src.(Imm); isImm && isLarge(imm.Value) {
				return []AsmInstruction{
					AsmMov{
						Type: &asmtype.QuadWord{},
						Src:  imm,
						Dst:  Register{Reg: R10},
					},
					AsmBinary{
						Op:   ast.Op,
						Type: &asmtype.QuadWord{},
						Src:  Register{Reg: R10},
						Dst:  ast.Dst,
					},
				}
			}
		}

		// Handle memory operands for Add/Sub
		if ast.Op == Add || ast.Op == Sub {
			if srcStack, isSrcStack := ast.Src.(Stack); isSrcStack {
				if dstStack, isDstStack := ast.Dst.(Stack); isDstStack {
					return []AsmInstruction{
						AsmMov{
							Type: ast.Type,
							Src:  srcStack,
							Dst:  Register{Reg: R10},
						},
						AsmBinary{
							Op:   ast.Op,
							Type: ast.Type,
							Src:  Register{Reg: R10},
							Dst:  dstStack,
						},
					}
				}
			}
		}

		// Handle Mult with memory destination
		if ast.Op == Mult {
			if dstStack, isDstStack := ast.Dst.(Stack); isDstStack {
				return []AsmInstruction{
					AsmMov{
						Type: ast.Type,
						Src:  dstStack,
						Dst:  Register{Reg: R11},
					},
					AsmBinary{
						Op:   Mult,
						Type: ast.Type,
						Src:  ast.Src,
						Dst:  Register{Reg: R11},
					},
					AsmMov{
						Type: ast.Type,
						Src:  Register{Reg: R11},
						Dst:  dstStack,
					},
				}
			}
		}
		return []AsmInstruction{ast}

	case Cmp:
		// Handle memory operands
		if srcStack, isSrcStack := ast.Src.(Stack); isSrcStack {
			if dstStack, isDstStack := ast.Dst.(Stack); isDstStack {
				return []AsmInstruction{
					AsmMov{
						Type: ast.Type,
						Src:  srcStack,
						Dst:  Register{Reg: R10},
					},
					Cmp{
						Type: ast.Type,
						Src:  Register{Reg: R10},
						Dst:  dstStack,
					},
				}
			}
		}

		// Handle immediate operands
		if srcImm, isSrcImm := ast.Src.(Imm); isSrcImm {
			if dstImm, isDstImm := ast.Dst.(Imm); isDstImm {
				if _, isQuadWord := ast.Type.(*asmtype.QuadWord); isQuadWord && isLarge(srcImm.Value) {
					return []AsmInstruction{
						AsmMov{
							Type: &asmtype.QuadWord{},
							Src:  srcImm,
							Dst:  Register{Reg: R10},
						},
						AsmMov{
							Type: &asmtype.QuadWord{},
							Src:  dstImm,
							Dst:  Register{Reg: R11},
						},
						Cmp{
							Type: &asmtype.QuadWord{},
							Src:  Register{Reg: R10},
							Dst:  Register{Reg: R11},
						},
					}
				}
			}
			if _, isQuadWord := ast.Type.(*asmtype.QuadWord); isQuadWord && isLarge(srcImm.Value) {
				return []AsmInstruction{
					AsmMov{
						Type: &asmtype.QuadWord{},
						Src:  srcImm,
						Dst:  Register{Reg: R10},
					},
					Cmp{
						Type: &asmtype.QuadWord{},
						Src:  Register{Reg: R10},
						Dst:  ast.Dst,
					},
				}
			}
		}

		// Handle immediate destination
		if dstImm, isDstImm := ast.Dst.(Imm); isDstImm {
			return []AsmInstruction{
				AsmMov{
					Type: ast.Type,
					Src:  dstImm,
					Dst:  Register{Reg: R11},
				},
				Cmp{
					Type: ast.Type,
					Src:  ast.Src,
					Dst:  Register{Reg: R11},
				},
			}
		}
		return []AsmInstruction{ast}

	case Push:
		// Handle large immediate
		if imm, isImm := ast.Op.(Imm); isImm && isLarge(imm.Value) {
			return []AsmInstruction{
				AsmMov{
					Type: &asmtype.QuadWord{},
					Src:  imm,
					Dst:  Register{Reg: R10},
				},
				Push{
					Op: Register{Reg: R10},
				},
			}
		}
		return []AsmInstruction{ast}

	default:
		return []AsmInstruction{instr}
	}
}

func (f *FixUpPassGen) FixUpInFn(fn AsmFnDef) AsmFnDef {
	stackFrameSize := f.symbolTable.Get(fn.Ident).StackFrameSize
	stackBytes := roundingutil.RoundAwayFromZero(16, stackFrameSize)
	stkByteOp := Imm{Value: int64(stackBytes)}
	allocStack := AsmBinary{
		Op:   Sub,
		Type: &asmtype.QuadWord{},
		Src:  stkByteOp,
		Dst:  Register{Reg: SP},
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
	return AsmProgram{AsmFnDef: asmFnDefs, AsmExternFn: program.AsmExternFn}
}
