package asmtype

type AsmType interface {
	asmtype()
}

type LongWord struct{}

func (l *LongWord) asmtype() {}

type QuadWord struct{}

func (l *QuadWord) asmtype() {}
