package gen

import (
	"fmt"
	"os"
)

type QBEEmitter struct {
	WriteFile  *os.File
	Irs        []IR
	LocalCount uint
}

func NewQBEEmitter(file *os.File, Irs []IR) Emitter {
	return QBEEmitter{
		WriteFile:  file,
		Irs:        Irs,
		LocalCount: 0,
	}
}

func (x QBEEmitter) Emit() {
	x.starter()
	for _, ir := range x.Irs {
		switch irtype := ir.(type) {
		case *IRMov:
			x.WriteFile.WriteString("    # push instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString(fmt.Sprintf("  %%tmp.%d =w add %d, 0", x.LocalCount, irtype.Dst))
			x.WriteFile.WriteString("    \n")
			// fmt.Println(irtype.Ir())
			x.LocalCount += 1
		case *IRReturn:
			x.WriteFile.WriteString("    # return instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString("    \n")
			// fmt.Println(irtype.Ir())
		case *IRPrint:
			x.WriteFile.WriteString("    # print instruction \n")
			x.WriteFile.WriteString("    # not implemented \n")
			x.WriteFile.WriteString("    \n")
		}
	}
	x.ending()
}

func (x QBEEmitter) starter() {
	x.WriteFile.WriteString("export function w $main() {\n")
	x.WriteFile.WriteString("@start\n")
}

func (x QBEEmitter) ending() {
	x.WriteFile.WriteString("  ret 0\n")
	x.WriteFile.WriteString("}\n")
}
