package gen

import (
	"fmt"
	"os"
)

type QBEEmitter struct {
	WriteFile *os.File
	Irs       []IR
}

func NewQBEEmitter(file *os.File, Irs []IR) Emitter {
	return QBEEmitter{
		WriteFile: file,
		Irs:       Irs,
	}
}

func (x QBEEmitter) Emit() {
	x.starter()
	for _, ir := range x.Irs {
		switch irtype := ir.(type) {
		case *IRPush:
			x.WriteFile.WriteString("    # push instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString("    \n")
			fmt.Println(irtype.Ir())
		case *IRReturn:
			x.WriteFile.WriteString("    # return instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString("    \n")
			fmt.Println(irtype.Ir())
		case *IRPrint:
			x.WriteFile.WriteString("    # print instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString("    \n")
		}
	}
	x.ending()
}

func (x QBEEmitter) starter() {
	x.WriteFile.WriteString("data $str = { b \"hello world\", b 0 }\n")
	x.WriteFile.WriteString("export function w $main() {\n")
	x.WriteFile.WriteString("@start\n")
	x.WriteFile.WriteString("%r =w call $puts(l $str)\n")
	x.WriteFile.WriteString("   ret 0\n")
	x.WriteFile.WriteString("}\n")
}

func (x QBEEmitter) ending() {
}
