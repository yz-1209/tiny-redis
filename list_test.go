package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	l := NewList(ListType{EqualFunc: StrEqual})
	assert.Equal(t, 0, l.length)

	l.Append(NewObject(String, "4"))
	l.DelNode(l.First())

	l.Append(NewObject(String, "1"))
	l.Append(NewObject(String, "2"))
	l.Append(NewObject(String, "3"))
	assert.Equal(t, 3, l.length)
	assert.Equal(t, "1", l.First().Val.StrVal())
	assert.Equal(t, "3", l.Last().Val.StrVal())

	o := NewObject(String, "0")
	l.LPush(o)
	assert.Equal(t, 4, l.length)
	assert.Equal(t, "0", l.First().Val.StrVal())

	l.LPush(NewObject(String, "-1"))
	assert.Equal(t, 5, l.length)

	n := l.Find(o)
	assert.Equal(t, o, n.Val)

	l.Delete(o)
	assert.Equal(t, 4, l.length)
	n = l.Find(o)
	assert.Nil(t, n)

	l.DelNode(l.First())
	assert.Equal(t, 3, l.length)
	assert.Equal(t, "1", l.First().Val.StrVal())

	l.DelNode(l.Last())
	assert.Equal(t, 2, l.length)
	assert.Equal(t, "2", l.Last().Val.StrVal())
}
