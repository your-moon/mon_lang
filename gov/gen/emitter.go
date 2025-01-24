package gen

import (
	"fmt"
	"os"
)

type MachineTarget string

const (
	X86 MachineTarget = "x64"
	QBE MachineTarget = "qbe"
)

type Emitter struct {
	WriteFile *os.File
	Irs       []IR
	Target    MachineTarget
}

func NewEmitter(file *os.File, Irs []IR, target MachineTarget) Emitter {
	return Emitter{
		WriteFile: file,
		Irs:       Irs,
		Target:    target,
	}
}

func (e *Emitter) Starter() {
	if e.Target == QBE {

		e.WriteFile.WriteString("data $str = { b \"hello world\", b 0 }\n")
		e.WriteFile.WriteString("export function w $main() {\n")
		e.WriteFile.WriteString("@start\n")
		e.WriteFile.WriteString("%r =w call $puts(l $str)\n")
		e.WriteFile.WriteString("   ret 0\n")
		e.WriteFile.WriteString("}\n")
	} else {
		e.WriteFile.WriteString("    .globl _main\n")
		e.WriteFile.WriteString("_main:\n")
		//prologue
		e.WriteFile.WriteString("    # prologue start\n")
		e.WriteFile.WriteString("    pushq %rbp\n")
		e.WriteFile.WriteString("    movq %rsp, %rbp\n")
		e.WriteFile.WriteString("    subq $8, %rsp\n")
		e.WriteFile.WriteString("    # prologue end\n")
		e.WriteFile.WriteString("\n")
	}
}
func (e *Emitter) Ending() {
	if e.Target == X86 {
		//epilogue
		e.WriteFile.WriteString("    # epilogue start\n")
		e.WriteFile.WriteString("    movq %rbp, %rsp\n")
		e.WriteFile.WriteString("    popq %rbp\n")
		e.WriteFile.WriteString("    ret\n")
		e.WriteFile.WriteString("    # epilogue end\n")
	}
}

func (e *Emitter) Emit() error {
	e.Starter()
	for _, ir := range e.Irs {
		switch irtype := ir.(type) {
		case *IRPush:
			if e.Target == X86 {
				e.WriteFile.WriteString("    # push instruction\n")
				e.WriteFile.WriteString("    movl $2, -4(%rbp)\n")
				e.WriteFile.WriteString("    push %rax\n")
			}
			fmt.Println(irtype.Ir())
		case *IRReturn:
			if e.Target == X86 {
				e.WriteFile.WriteString("    # return instruction\n")
				e.WriteFile.WriteString("    pop %rax\n")
				e.WriteFile.WriteString("    ret\n")
			}
			fmt.Println(irtype.Ir())
		}
	}
	e.Ending()
	return nil
}
