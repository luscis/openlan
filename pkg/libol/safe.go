package libol

import "sync"

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

	if sm.size != 0 && len(sm.data) >= sm.size {
		return NewErr("SafeStrStr.Set already full")
	}
	sm.data[k] = v
	return nil
}

func (sm *SafeStrStr) Set(k, v string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.size != 0 && len(sm.data) >= sm.size {
		return NewErr("SafeStrStr.Set already full")
	}
	if _, ok := sm.data[k]; !ok {
		sm.data[k] = v
	}
	return nil
}

func (sm *SafeStrStr) Del(k string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if _, ok := sm.data[k]; ok {
		delete(sm.data, k)
	}
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
	data map[string]interface{}
	lock sync.RWMutex
}

func NewSafeStrMap(size int) *SafeStrMap {
	calSize := size
	if calSize == 0 {
		calSize = 128
	}
	return &SafeStrMap{
		size: size,
		data: make(map[string]interface{}, calSize),
	}
}

func (sm *SafeStrMap) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return len(sm.data)
}

func (sm *SafeStrMap) add(k string, v interface{}) error {
	if sm.size != 0 && len(sm.data) >= sm.size {
		return NewErr("SafeStrMap.Set already full")
	}
	if _, ok := sm.data[k]; !ok {
		sm.data[k] = v
	}
	return nil
}

func (sm *SafeStrMap) Set(k string, v interface{}) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	return sm.add(k, v)
}

func (sm *SafeStrMap) Mod(k string, v interface{}) error {
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

	if _, ok := sm.data[k]; ok {
		delete(sm.data, k)
	}
}

func (sm *SafeStrMap) Clear() {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.data = make(map[string]interface{}, sm.size)
}

func (sm *SafeStrMap) Get(k string) interface{} {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return sm.data[k]
}

func (sm *SafeStrMap) GetEx(k string) (interface{}, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	v, ok := sm.data[k]
	return v, ok
}

func (sm *SafeStrMap) Iter(proc func(k string, v interface{})) int {
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
	data interface{}
	lock sync.RWMutex
}

func NewSafeVar() *SafeVar {
	return &SafeVar{}
}

func (sv *SafeVar) Set(v interface{}) {
	sv.lock.Lock()
	defer sv.lock.Unlock()
	sv.data = v
}

func (sv *SafeVar) Get() interface{} {
	sv.lock.RLock()
	defer sv.lock.RUnlock()
	return sv.data
}

func (sv *SafeVar) GetWithFunc(proc func(v interface{})) {
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
	if _, ok := s.data[k]; ok {
		s.data[k] += v
	} else {
		s.data[k] = v
	}
}

func (s *SafeStrInt64) Copy(dst map[string]int64) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.data {
		dst[k] = v
	}
}

func (s *SafeStrInt64) Data() map[string]int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	dst := make(map[string]int64, 32)
	for k, v := range s.data {
		dst[k] = v
	}
	return dst
}
