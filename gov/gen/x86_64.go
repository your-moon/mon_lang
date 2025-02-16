package gen

import (
	"fmt"
	"os"

	"github.com/your-moon/mn_compiler_go_version/utfconvert"
)

type X86_64Emitter struct {
	WriteFile *os.File
	Irs       []IR
}

func NewX86Emitter(file *os.File, Irs []IR) Emitter {
	return X86_64Emitter{
		WriteFile: file,
		Irs:       Irs,
	}
}

func (x X86_64Emitter) Emit() {
	// x.starter()
	for _, ir := range x.Irs {
		switch irtype := ir.(type) {
		case *IRFn:
			x.WriteFile.WriteString("    # fn stmt construct\n")
			x.WriteFile.WriteString(fmt.Sprintf("    .globl _%s\n", utfconvert.UtfConvert(irtype.Name)))
			x.WriteFile.WriteString(fmt.Sprintf("_%s:\n", utfconvert.UtfConvert(irtype.Name)))
			x.WriteFile.WriteString("    pushq %rbp\n")
			x.WriteFile.WriteString("    movq %rsp, %rbp\n")

		// case *IRMov:
		// 	x.WriteFile.WriteString("    # push instruction\n")
		// 	x.WriteFile.WriteString(fmt.Sprintf("    movl $%d, -4(%%rbp)\n", irtype.Value))
		// 	x.WriteFile.WriteString("    pushq rbp\n")
		// fmt.Println(irtype.Ir())
		case *IRReturn:
			x.WriteFile.WriteString("    # return instruction\n")
			x.WriteFile.WriteString("    movq %rbp, %rsp\n")
			x.WriteFile.WriteString("    popq %rbp\n")
			x.WriteFile.WriteString("    ret\n")
			// fmt.Println(irtype.Ir())
		case *IRPrint:
			x.WriteFile.WriteString("    # print instruction # not implemented \n")
			x.WriteFile.WriteString("    \n")
		}
	}
	// x.exit()
	// x.ending()
}

func (x X86_64Emitter) exit() {
	//exit
	x.WriteFile.WriteString("    movq %rbp, %rsp\n")
	x.WriteFile.WriteString("    popq %rbp\n")
	x.WriteFile.WriteString("    ret\n")
	x.WriteFile.WriteString("\n")
}

func (x X86_64Emitter) starter() {
	x.WriteFile.WriteString("    .globl _main\n")
	x.WriteFile.WriteString("_main:\n")
	//prologue
	x.WriteFile.WriteString("    # prologue start\n")
	x.WriteFile.WriteString("    pushq %rbp\n")
	x.WriteFile.WriteString("    movq %rsp, %rbp\n")
	x.WriteFile.WriteString("    subq $8, %rsp\n")
	x.WriteFile.WriteString("    # prologue end\n")
	x.WriteFile.WriteString("\n")
}

func (x X86_64Emitter) ending() {
	//epilogue
	x.WriteFile.WriteString("    # epilogue start\n")
	x.WriteFile.WriteString("    movq %rbp, %rsp\n")
	x.WriteFile.WriteString("    popq %rbp\n")
	x.WriteFile.WriteString("    ret\n")
	x.WriteFile.WriteString("    # epilogue end\n")
}
