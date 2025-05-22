package symbols

import "github.com/your-moon/mn_compiler/mtypes"

type Entry struct {
	Type           mtypes.Type
	IsDefined      bool
	StackFrameSize int
}

type SymbolTable struct {
	Entries map[string]*Entry
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{Entries: make(map[string]*Entry)}
}

func (s *SymbolTable) Modify(name string, entry *Entry) {
	s.Entries[name] = entry
}

func (s *SymbolTable) AddFn(t *mtypes.FnType, name string, isDefined bool) *Entry {
	entry := &Entry{
		Type:           t,
		IsDefined:      isDefined,
		StackFrameSize: 0,
	}
	s.Entries[name] = entry
	return entry
}

func (s *SymbolTable) AddVar(t mtypes.Type, name string) *Entry {
	entry := &Entry{
		Type:           t,
		IsDefined:      false,
		StackFrameSize: 0,
	}
	s.Entries[name] = entry
	return entry
}

func (s *SymbolTable) Get(name string) *Entry {
	entry, ok := s.Entries[name]
	if !ok {
		return nil
	}
	return entry
}

func (s *SymbolTable) GetOptional(name string) *Entry {
	entry, ok := s.Entries[name]
	if !ok {
		return nil
	}
	return entry
}

func (s *SymbolTable) IsDefined(name string) bool {
	entry := s.Get(name)
	return entry != nil && entry.IsDefined
}

func (s *SymbolTable) SetBytesRequired(name string, bytes int) {
	entry := s.Get(name)
	entry.StackFrameSize = bytes
}

func (s *SymbolTable) Bindings() [][2]interface{} {
	result := make([][2]interface{}, 0, len(s.Entries))
	for name, entry := range s.Entries {
		result = append(result, [2]interface{}{name, entry})
	}
	return result
}
