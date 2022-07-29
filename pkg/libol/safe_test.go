package libol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestSafeStrStr(t *testing.T) {
	m := NewSafeStrStr(1024)
	_ = m.Set("hi", "1")
	i := m.Get("hi")
	assert.Equal(t, i, "1", "be the same.")

	a := "3"
	_ = m.Set("hip", a)
	c := m.Get("hip")
	assert.Equal(t, c, a, "be the same.")
	assert.Equal(t, 2, m.Len(), "be the same.")

	for i := 0; i < 1024; i++ {
		_ = m.Set(fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
	}
	assert.Equal(t, 1024, m.Len(), "")
	fmt.Printf("TestSafeStrStr.size: %d\n", m.Len())
	for i := 0; i < 1024; i++ {
		m.Del(fmt.Sprintf("%d", i))
	}
	assert.Equal(t, 2, m.Len(), "")

	m.Del("hi")
	ii := m.Get("hi")
	assert.Equal(t, ii, "", "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")

	iii := m.Get("hello")
	assert.Equal(t, iii, "", "be the same.")
}

func TestZeroMapSet(t *testing.T) {
	m := make(map[string]int, 32)
	m["hi"] = 1
	i := m["hi"]
	assert.Equal(t, i, 1, "be the same.")

	m["hi"] = 3
	c := m["hi"]
	assert.Equal(t, c, 3, "be the same.")
	assert.Equal(t, 1, len(m), "be the same.")
}

func TestZeroSafeStrStrSet(t *testing.T) {
	m := NewSafeStrStr(0)
	_ = m.Set("hi", "1")
	i := m.Get("hi")
	assert.Equal(t, i, "1", "be the same.")

	_ = m.Set("hi", "3")
	c := m.Get("hi")
	assert.Equal(t, c, "1", "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")
}

func TestZeroSafeStrMapSet(t *testing.T) {
	m := NewSafeStrMap(0)
	_ = m.Set("hi", 1)
	i := m.Get("hi")
	assert.Equal(t, i, 1, "be the same.")

	_ = m.Set("hi", 3)
	c := m.Get("hi").(int)
	assert.Equal(t, c, 1, "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")
}

func TestZeroSafeStrStr(t *testing.T) {
	m := NewSafeStrStr(0)
	_ = m.Set("hi", "1")
	i := m.Get("hi")
	assert.Equal(t, i, "1", "be the same.")

	a := "3"
	_ = m.Set("hip", a)
	c := m.Get("hip")
	assert.Equal(t, c, a, "be the same.")
	assert.Equal(t, 2, m.Len(), "be the same.")

	for i := 0; i < 1024; i++ {
		_ = m.Set(fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
	}
	assert.Equal(t, 1026, m.Len(), "")
	fmt.Printf("TestZeroSafeStrStr.size: %d\n", m.Len())
	for i := 0; i < 1024; i++ {
		m.Del(fmt.Sprintf("%d", i))
	}
	assert.Equal(t, 2, m.Len(), "")
	m.Del("hi")
	ii := m.Get("hi")
	assert.Equal(t, ii, "", "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")

	iii := m.Get("hello")
	assert.Equal(t, iii, "", "be the same.")
}

func TestZeroSafeStrStrIter(t *testing.T) {
	m := NewSafeStrStr(0)
	c := 0
	for i := 0; i < 10; i++ {
		c += i
		_ = m.Set(fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
	}
	ct := 0
	m.Iter(func(k string, v string) {
		i, _ := strconv.Atoi(v)
		ct += i
	})
	assert.Equal(t, ct, c, "be the same")

	ms := NewSafeStrMap(0)
	cm := 0
	for i := 1024; i < 1024+1024; i++ {
		cm += i
		_ = ms.Set(fmt.Sprintf("%d", i), i)
	}
	cmt := 0
	ms.Iter(func(k string, v interface{}) {
		cmt += v.(int)
	})
	assert.Equal(t, cmt, cm, "be the same")
}

func TestSafeVar(t *testing.T) {
	v := NewSafeVar()
	a := 3
	c := 0
	v.Set(2)
	v.GetWithFunc(func(v interface{}) {
		c = a + v.(int)
	})
	assert.Equal(t, 5, c, "")
}

func BenchmarkMapGet(b *testing.B) {
	m := make(map[string]int, 2)
	m["hi"] = 2

	for i := 0; i < b.N; i++ {
		v := m["hi"]
		assert.Equal(b, v, 2, "")
	}
}

func BenchmarkMapGetWithLock(b *testing.B) {
	m := make(map[string]int, 2)
	m["hi"] = 2
	lock := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		lock.RLock()
		v := m["hi"]
		lock.RUnlock()
		assert.Equal(b, v, 2, "")
	}
}

func BenchmarkSafeStrStrGet(b *testing.B) {
	m := NewSafeStrStr(2)
	_ = m.Set("hi", "2")

	for i := 0; i < b.N; i++ {
		v := m.Get("hi")
		assert.Equal(b, v, "2", "")
	}
}

func BenchmarkSafeStrMapGet(b *testing.B) {
	m := NewSafeStrMap(2)
	_ = m.Set("hi", 2)

	for i := 0; i < b.N; i++ {
		v := m.Get("hi").(int)
		assert.Equal(b, v, 2, "")
	}
}
