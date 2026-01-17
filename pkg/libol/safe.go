package libol

import (
	"maps"
	"sync"
)

//m := NewSafeStrStr(1024)
//m.Set("hi", "1")
//a := "3"
//m.Set("hip", a)
//c := m.Get("hip")
//fmt.Printf("%s\n%s\n", m, c)

type SafeStrStr struct {
	size int
	data map[string]string
	lock sync.RWMutex
}

func NewSafeStrStr(size int) *SafeStrStr {
	calSize := size
	if calSize == 0 {
		calSize = 128
	}
	return &SafeStrStr{
		size: size,
		data: make(map[string]string, calSize),
	}
}

func (sm *SafeStrStr) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return len(sm.data)
}

func (sm *SafeStrStr) Reset(k, v string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.size == 0 || len(sm.data) < sm.size {
		sm.data[k] = v
		return nil
	}
	return NewErr("SafeStrStr.Set already full")
}

func (sm *SafeStrStr) Set(k, v string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.size == 0 || len(sm.data) < sm.size {
		if _, ok := sm.data[k]; !ok {
			sm.data[k] = v
		}
		return nil
	}

	return NewErr("SafeStrStr.Set already full")
}

func (sm *SafeStrStr) Del(k string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	delete(sm.data, k)
}

func (sm *SafeStrStr) Get(k string) string {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return sm.data[k]
}

func (sm *SafeStrStr) GetEx(k string) (string, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	v, ok := sm.data[k]
	return v, ok
}

func (sm *SafeStrStr) Iter(proc func(k, v string)) int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	count := 0
	for k, u := range sm.data {
		if k != "" {
			proc(k, u)
			count += 1
		}
	}
	return count
}

type SafeStrMap struct {
	size int
	data map[string]any
	lock sync.RWMutex
}

func NewSafeStrMap(size int) *SafeStrMap {
	calSize := size
	if calSize == 0 {
		calSize = 128
	}
	return &SafeStrMap{
		size: size,
		data: make(map[string]any, calSize),
	}
}

func (sm *SafeStrMap) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return len(sm.data)
}

func (sm *SafeStrMap) Full() bool {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	if sm.size == 0 || len(sm.data) < sm.size {
		return false
	}
	return true
}

func (sm *SafeStrMap) add(k string, v any) error {
	if sm.size == 0 || len(sm.data) < sm.size {
		if _, ok := sm.data[k]; !ok {
			sm.data[k] = v
		}
		return nil
	}
	return NewErr("SafeStrMap.Set already full")
}

func (sm *SafeStrMap) Set(k string, v any) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	return sm.add(k, v)
}

func (sm *SafeStrMap) Mod(k string, v any) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if _, ok := sm.data[k]; !ok {
		return sm.add(k, v)
	}
	sm.data[k] = v
	return nil
}

func (sm *SafeStrMap) Del(k string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	delete(sm.data, k)
}

func (sm *SafeStrMap) Clear() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sm.data = make(map[string]any, sm.size)
}

func (sm *SafeStrMap) Get(k string) any {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return sm.data[k]
}

func (sm *SafeStrMap) GetEx(k string) (any, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	v, ok := sm.data[k]
	return v, ok
}

func (sm *SafeStrMap) Iter(proc func(k string, v any)) int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	count := 0
	for k, u := range sm.data {
		if u != nil {
			proc(k, u)
			count += 1
		}
	}
	return count
}

// a := SafeVar
// a.Set(0x01)
// a.Get().(int)

type SafeVar struct {
	data any
	lock sync.RWMutex
}

func NewSafeVar() *SafeVar {
	return &SafeVar{}
}

func (sv *SafeVar) Set(v any) {
	sv.lock.Lock()
	defer sv.lock.Unlock()

	sv.data = v
}

func (sv *SafeVar) Get() any {
	sv.lock.RLock()
	defer sv.lock.RUnlock()

	return sv.data
}

func (sv *SafeVar) GetWithFunc(proc func(v any)) {
	sv.lock.RLock()
	defer sv.lock.RUnlock()

	proc(sv.data)
}

type SafeStrInt64 struct {
	lock sync.RWMutex
	data map[string]int64
}

func NewSafeStrInt64() *SafeStrInt64 {
	return &SafeStrInt64{
		data: make(map[string]int64, 32),
	}
}

func (s *SafeStrInt64) Get(k string) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if v, ok := s.data[k]; ok {
		return v
	}
	return 0
}

func (s *SafeStrInt64) Set(k string, v int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[k] = v
}

func (s *SafeStrInt64) Add(k string, v int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[k] += v
}

func (s *SafeStrInt64) Copy(dst map[string]int64) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	maps.Copy(dst, s.data)
}

func (s *SafeStrInt64) Data() map[string]int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	dst := make(map[string]int64, 32)
	maps.Copy(dst, s.data)
	return dst
}
