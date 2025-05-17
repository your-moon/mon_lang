package symbols

import "github.com/your-moon/mn_compiler_go_version/mtypes"

type Entry struct {
	Type           mtypes.Type
	isDefined      bool
	StackFrameSize int
}

type SymbolTable struct {
	entries map[string]*Entry
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{entries: make(map[string]*Entry)}
}

func (s *SymbolTable) Modify(name string, entry *Entry) {
	s.entries[name] = entry
}

func (s *SymbolTable) AddFn(t *mtypes.FnType, name string, isDefined bool) *Entry {
	entry := &Entry{
		Type:           t,
		isDefined:      isDefined,
		StackFrameSize: 0,
	}
	s.entries[name] = entry
	return entry
}

func (s *SymbolTable) AddVar(t mtypes.Type, name string) *Entry {
	entry := &Entry{
		Type:           t,
		isDefined:      false,
		StackFrameSize: 0,
	}
	s.entries[name] = entry
	return entry
}

func (s *SymbolTable) Get(name string) *Entry {
	entry, ok := s.entries[name]
	if !ok {
		return nil
	}
	return entry
}

func (s *SymbolTable) GetOptional(name string) *Entry {
	entry, ok := s.entries[name]
	if !ok {
		return nil
	}
	return entry
}

func (s *SymbolTable) IsDefined(name string) bool {
	entry := s.Get(name)
	return entry != nil && entry.isDefined
}

func (s *SymbolTable) SetBytesRequired(name string, bytes int) {
	entry := s.Get(name)
	entry.StackFrameSize = bytes
}

func (s *SymbolTable) Bindings() [][2]interface{} {
	result := make([][2]interface{}, 0, len(s.entries))
	for name, entry := range s.entries {
		result = append(result, [2]interface{}{name, entry})
	}
	return result
}
