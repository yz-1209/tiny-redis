package main

import (
	"log"
	"math/rand"
)

const (
	InitSize    int64 = 8
	LoadFactor  int64 = 8
	DefaultStep int   = 1
)

type Entry struct {
	Key  *Obj
	Val  *Obj
	Next *Entry
}

type DictType struct {
	HashFunc  func(key *Obj) int64
	EqualFunc func(k1, k2 *Obj) bool
}

type HTable struct {
	DictType

	buckets []*Entry
	size    int64
	mask    int64
	used    int64
}

func NewHTable(n int64, dictType DictType) *HTable {
	return &HTable{
		DictType: dictType,
		buckets:  make([]*Entry, n),
		size:     n,
		mask:     n - 1,
	}
}

func (h *HTable) Insert(key, val *Obj) {
	hashcode := h.HashFunc(key)
	idx := h.mask & hashcode
	entry := &Entry{
		Key:  key,
		Val:  val,
		Next: h.buckets[idx],
	}
	h.buckets[idx] = entry
	h.used++
}

func (h *HTable) Lookup(key *Obj) *Entry {
	if h.used == 0 {
		return nil
	}

	hashcode := h.HashFunc(key)
	idx := h.mask & hashcode
	entry := h.buckets[idx]
	for entry != nil {
		if h.EqualFunc(entry.Key, key) {
			return entry
		}
		entry = entry.Next
	}
	return nil
}

func (h *HTable) Pop(key *Obj) *Entry {
	hashcode := h.HashFunc(key)
	idx := h.mask & hashcode
	dummy := &Entry{Next: h.buckets[idx]}
	prev, curr := dummy, dummy.Next
	for curr != nil && !h.EqualFunc(curr.Key, key) {
		prev, curr = curr, curr.Next
	}
	if curr == nil {
		return nil
	}
	prev.Next = curr.Next
	h.buckets[idx] = dummy.Next
	h.used--
	return curr
}

func (h *HTable) RandomGet() *Entry {
	if h.used == 0 {
		return nil
	}

	var bucketIndexes []int64
	for i := int64(0); i < h.size; i++ {
		if h.buckets[i] != nil {
			bucketIndexes = append(bucketIndexes, i)
		}
	}

	idx := bucketIndexes[rand.Int63n(int64(len(bucketIndexes)))]
	var listLen int64
	for p := h.buckets[idx]; p != nil; p = p.Next {
		listLen++
	}

	listIdx := rand.Int63n(listLen)
	p := h.buckets[idx]
	for i := int64(0); i < listIdx; i++ {
		p = p.Next
	}
	return p
}

type Dict struct {
	DictType
	tab1        *HTable
	tab2        *HTable
	resizingIdx int64
}

func NewDict(dictType DictType) *Dict {
	return &Dict{
		DictType:    dictType,
		resizingIdx: -1,
	}
}

func (d *Dict) startResizing() {
	if d.tab2 != nil {
		log.Panic("dict tab2 is't nil")
	}

	d.tab2 = d.tab1
	d.tab1 = NewHTable(d.tab2.size*2, d.DictType)
	d.resizingIdx = 0
}

func (d *Dict) resizing() {
	d.resizeStep(DefaultStep)
}

func (d *Dict) resizeStep(step int) {
	for step > 0 {
		if d.tab2 == nil {
			return
		}

		for d.resizingIdx < d.tab2.size && d.tab2.buckets[d.resizingIdx] == nil {
			d.resizingIdx++
		}

		if d.resizingIdx == d.tab2.size {
			d.tab2 = nil
			return
		}

		head := d.tab2.buckets[d.resizingIdx]
		for head != nil {
			next := head.Next
			d.tab1.Insert(head.Key, head.Val)
			d.tab2.used--
			head = next
		}
		d.tab2.buckets[d.resizingIdx] = nil
		d.resizingIdx++
		step--
	}
}

func (d *Dict) Insert(key, val *Obj) {
	if d.tab1 == nil {
		d.tab1 = NewHTable(InitSize, d.DictType)
	}

	d.resizing()

	entry := d.lookup(key)
	if entry != nil {
		entry.Val.DecrRefCount()
		entry.Val = val
		val.IncrRefCount()
		return
	}

	d.tab1.Insert(key, val)

	key.IncrRefCount()
	val.IncrRefCount()

	if d.tab2 == nil {
		factor := (d.tab1.used - 1) / d.tab1.size
		if factor >= LoadFactor {
			d.startResizing()
		}
	}
}

func (d *Dict) lookup(key *Obj) *Entry {
	for _, tab := range []*HTable{d.tab1, d.tab2} {
		if tab != nil {
			entry := tab.Lookup(key)
			if entry != nil {
				return entry
			}
		}
	}
	return nil
}

func (d *Dict) Lookup(key *Obj) *Entry {
	d.resizing()
	return d.lookup(key)
}

func (d *Dict) Pop(key *Obj) *Entry {
	d.resizing()

	for _, tab := range []*HTable{d.tab1, d.tab2} {
		if tab != nil {
			entry := tab.Lookup(key)
			if entry != nil {
				_ = tab.Pop(key)
				entry.Key.DecrRefCount()
				entry.Val.DecrRefCount()
				return entry
			}
		}
	}
	return nil
}

func (d *Dict) RandomGet() *Entry {
	if d.tab1 == nil {
		return nil
	}

	d.resizing()

	tab := d.tab1
	if d.tab2 != nil {
		cnt := d.tab1.used + d.tab2.used
		if rand.Int63n(cnt) >= d.tab1.used {
			tab = d.tab2
		}
	}

	return tab.RandomGet()
}

func (d *Dict) KeyCount() int64 {
	if d.tab1 == nil {
		return 0
	}

	cnt := d.tab1.used
	if d.tab2 != nil {
		cnt += d.tab2.used
	}
	return cnt
}
