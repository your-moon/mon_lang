package stringpool

import (
	"fmt"
	"sync"
)

type StringPool struct {
	mutex       sync.RWMutex
	strings     map[string]int
	stringCount int
}

var globalPool = NewStringPool()

func NewStringPool() *StringPool {
	return &StringPool{
		strings:     make(map[string]int),
		stringCount: 0,
	}
}

func (p *StringPool) Intern(value string) int {
	p.mutex.RLock()
	id, exists := p.strings[value]
	p.mutex.RUnlock()

	if exists {
		return id
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if id, exists = p.strings[value]; exists {
		return id
	}

	id = p.stringCount
	p.strings[value] = id
	p.stringCount++
	return id
}

func (p *StringPool) GetString(id int) (string, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for str, strID := range p.strings {
		if strID == id {
			return str, true
		}
	}
	return "", false
}

func (p *StringPool) GetID(value string) (int, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	id, exists := p.strings[value]
	return id, exists
}

func (p *StringPool) GetLabel(value string) string {
	id := p.Intern(value)
	return fmt.Sprintf(".LC%d", id)
}

func (p *StringPool) GetAllStrings() map[string]int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	result := make(map[string]int, len(p.strings))
	for k, v := range p.strings {
		result[k] = v
	}
	return result
}

func Intern(value string) int {
	return globalPool.Intern(value)
}

func GetString(id int) (string, bool) {
	return globalPool.GetString(id)
}

func GetID(value string) (int, bool) {
	return globalPool.GetID(value)
}

func GetLabel(value string) string {
	return globalPool.GetLabel(value)
}

func GetAllStrings() map[string]int {
	return globalPool.GetAllStrings()
}
