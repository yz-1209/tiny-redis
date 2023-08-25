package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDict(t *testing.T) {
	dict := NewDict(DictType{HashFunc: StrHash, EqualFunc: StrEqual})
	entry := dict.RandomGet()
	assert.Nil(t, entry)

	k1 := NewObject(String, "k1")
	v1 := NewObject(String, "v1")
	dict.Insert(k1, v1)

	entry = dict.Lookup(k1)
	assert.Equal(t, k1, entry.Key)
	assert.Equal(t, v1, entry.Val)
	assert.Equal(t, 2, k1.refCount)
	assert.Equal(t, 2, v1.refCount)

	dict.Pop(k1)
	entry = dict.Lookup(k1)
	assert.Nil(t, entry)
	assert.Equal(t, 1, k1.refCount)
	assert.Equal(t, 1, v1.refCount)

	dict.Insert(k1, v1)
	entry = dict.Lookup(k1)
	assert.Equal(t, v1, entry.Val)
	v2 := NewObject(String, "v2")
	dict.Insert(k1, v2)
	entry = dict.Lookup(k1)
	assert.Equal(t, v2, entry.Val)
	assert.Equal(t, 2, v2.refCount)
	assert.Equal(t, 1, v1.refCount)
}

func TestRehash(t *testing.T) {
	dict := NewDict(DictType{HashFunc: StrHash, EqualFunc: StrEqual})
	entry := dict.RandomGet()
	assert.Nil(t, entry)

	num := int(InitSize * LoadFactor)
	for i := 0; i < num; i++ {
		key := NewObject(String, fmt.Sprintf("k%v", i))
		val := NewObject(String, fmt.Sprintf("v%v", i))
		dict.Insert(key, val)
	}

	assert.Nil(t, dict.tab2)

	key := NewObject(String, fmt.Sprintf("k%v", num))
	val := NewObject(String, fmt.Sprintf("v%v", num))
	dict.Insert(key, val)

	assert.NotNil(t, dict.tab2)
	assert.Equal(t, int64(0), dict.resizingIdx)
	assert.Equal(t, InitSize, dict.tab2.size)
	assert.Equal(t, InitSize*2, dict.tab1.size)

	for i := 0; i <= int(InitSize); i++ {
		dict.RandomGet()
	}

	assert.Nil(t, dict.tab2)
	assert.Equal(t, InitSize*2, dict.tab1.size)
	for i := 0; i <= num; i++ {
		key := NewObject(String, fmt.Sprintf("k%v", i))
		entry := dict.Lookup(key)
		assert.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("v%v", i), entry.Val.StrVal())
	}
}
