package main

import (
    "strconv"
    "hash/fnv"
)

type ObjType uint8

const (
	String ObjType = iota
)

type Obj struct {
	Type     ObjType
	Val      any
	refCount int
}

func NewObject(type_ ObjType, ptr any) *Obj {
	return &Obj{
		Type:     type_,
		Val:      ptr,
		refCount: 1,
	}
}

func NewObjectInt(val int64) *Obj {
	return &Obj{
		Type:     String,
		Val:      strconv.FormatInt(val, 10),
		refCount: 1,
	}
}

func (o *Obj) IntVal() int64 {
	if o.Type != String {
		return 0
	}
	val, _ := strconv.ParseInt(o.Val.(string), 10, 64)
	return val
}

func (o *Obj) StrVal() string {
	if o.Type != String {
		return ""
	}
	return o.Val.(string)
}

func (o *Obj) IncrRefCount() {
	o.refCount++
}

func (o *Obj) DecrRefCount() {
	o.refCount--
	if o.refCount == 0 {
		o.Val = nil
	}
}

func StrHash(key *Obj) int64 {
	if key.Type != String {
		return 0
	}

	hash := fnv.New64()
	hash.Write([]byte(key.StrVal()))
	return int64(hash.Sum64())
}

func StrEqual(x, y *Obj) bool {
	if x.Type != String || y.Type != String {
		return false
	}

	return x.StrVal() == y.StrVal()
}
