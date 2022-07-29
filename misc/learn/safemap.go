package main

import (
	"fmt"
	"sync"
)

type SMap struct {
	Data map[interface{}]interface{}
	Lock sync.RWMutex
}

func NewSMap(size int) *SMap {
	this := &SMap{
		Data: make(map[interface{}]interface{}, size),
	}
	return this
}

func (sm *SMap) Set(k interface{}, v interface{}) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Data[k] = v
}

func (sm *SMap) Get(k interface{}) interface{} {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Data[k]
}

func (sm *SMap) GetEx(k string) (interface{}, bool) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	v, ok := sm.Data[k]
	return v, ok
}

func main() {
	m := NewSMap(1024)
	m.Set("hi", 1)
	fmt.Println(m)
	m.Set("hello", &m)

	fmt.Println(m)
	a := m.Get("hi").(int)
	a = 2
	fmt.Println(a)
	m.Set("hip", &a)
	fmt.Println(m)

	b := m.Get("hip").(*int)
	*b = 3
	fmt.Println(*b)
	c := m.Get("hip").(*int)
	fmt.Println(m)
	fmt.Println(*c)
}
