package store

import (
	"github.com/parashmaity/fleare/internal/comm"
)

type Memory struct {
	Mem map[string]*comm.Object
}

func NewMemory() *Memory {
	return &Memory{
		Mem: make(map[string]*comm.Object),
	}
}

func (m *Memory) Length() int32 {
	return int32(len(m.Mem))
}

func (m *Memory) Get(key string) (*comm.Object, error) {
	obj, ok := m.Mem[key]
	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (m *Memory) Set(key string, value []byte) error {
	obj := &comm.Object{Value: value}
	// fmt.Println("mem", obj, string(value))
	m.Mem[key] = obj
	return nil
}

func (m *Memory) Delete(key string) (bool, error) {
	if m.Mem[key] != nil {
		delete(m.Mem, key)
		return true, nil
	}
	return false, nil
}
