package asmsymbol

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/code_gen/asmtype"
)

// Entry is the interface implemented by both Fun and Obj
type Entry interface {
	isEntry()
}

// Fun represents a function entry
type Fun struct {
	IsDefined     bool
	BytesRequired int
}

func (f *Fun) isEntry() {}

// Obj represents a variable/object entry
type Obj struct {
	Type     asmtype.AsmType
	IsStatic bool
}

func (o *Obj) isEntry() {}

// SymbolTable represents the symbol table
type SymbolTable struct {
	entries map[string]Entry
}

func NewAsmSymbolTable() *SymbolTable {
	return &SymbolTable{entries: make(map[string]Entry)}
}

func (s *SymbolTable) AddFun(name string, defined bool) {
	s.entries[name] = &Fun{
		IsDefined:     defined,
		BytesRequired: 0,
	}
}

func (s *SymbolTable) AddVar(name string, t asmtype.AsmType, isStatic bool) {
	s.entries[name] = &Obj{
		Type:     t,
		IsStatic: isStatic,
	}
}

func (s *SymbolTable) SetBytesRequired(name string, bytes int) error {
	entry, ok := s.entries[name]
	if !ok {
		return fmt.Errorf("internal error: function %q is not defined", name)
	}
	fun, ok := entry.(*Fun)
	if !ok {
		return fmt.Errorf("internal error: %q is not a function", name)
	}
	fun.IsDefined = true
	fun.BytesRequired = bytes
	return nil
}

func (s *SymbolTable) GetBytesRequired(name string) (int, error) {
	entry, ok := s.entries[name]
	if !ok {
		return 0, fmt.Errorf("internal error: function %q not found", name)
	}
	fun, ok := entry.(*Fun)
	if !ok {
		return 0, fmt.Errorf("internal error: %q is not a function", name)
	}
	return fun.BytesRequired, nil
}

func (s *SymbolTable) GetSize(name string) (int, error) {
	entry, ok := s.entries[name]
	if !ok {
		return 0, fmt.Errorf("internal error: symbol %q not found", name)
	}
	obj, ok := entry.(*Obj)
	if !ok {
		return 0, fmt.Errorf("internal error: %q is a function, not an object", name)
	}
	switch obj.Type.(type) {
	case *asmtype.LongWord:
		return 4, nil
	case *asmtype.QuadWord:
		return 8, nil
	case *asmtype.StringType:
		return 8, nil // Strings are pointers, so they're 8 bytes on 64-bit systems
	default:
		return 0, fmt.Errorf("internal error: unknown asm type for %q", name)
	}
}

func (s *SymbolTable) GetAlignment(name string) (int, error) {
	entry, ok := s.entries[name]
	if !ok {
		return 0, fmt.Errorf("internal error: symbol %q not found", name)
	}
	obj, ok := entry.(*Obj)
	if !ok {
		return 0, fmt.Errorf("internal error: %q is a function, not an object", name)
	}
	switch obj.Type.(type) {
	case *asmtype.LongWord:
		return 4, nil
	case *asmtype.QuadWord:
		return 8, nil
	case *asmtype.StringType:
		return 8, nil // Strings are pointers, so they're 8 bytes on 64-bit systems
	default:
		return 0, fmt.Errorf("internal error: unknown asm type for %q", name)
	}
}

func (s *SymbolTable) IsDefined(name string) (bool, error) {
	entry, ok := s.entries[name]
	if !ok {
		return false, fmt.Errorf("internal error: symbol %q not found", name)
	}
	fun, ok := entry.(*Fun)
	if !ok {
		return false, fmt.Errorf("internal error: %q is not a function", name)
	}
	return fun.IsDefined, nil
}
