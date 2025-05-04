package semanticanalysis

type Entry struct {
	Type           Type
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

func (s *SymbolTable) AddFn(t *FnType, name string, isDefined bool) *Entry {
	entry := &Entry{
		Type:           t,
		isDefined:      isDefined,
		StackFrameSize: 0,
	}
	s.entries[name] = entry
	return entry
}

func (s *SymbolTable) AddVar(t Type, name string) *Entry {
	entry := &Entry{
		Type:           t,
		isDefined:      false,
		StackFrameSize: 0,
	}
	s.entries[name] = entry
	return entry
}

func (s *SymbolTable) Get(name string) *Entry {
	return s.entries[name]
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
