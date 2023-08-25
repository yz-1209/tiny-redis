package main

import (
    "time"
)

type GodisDB struct {
	data   *Dict
	expire *Dict
}

func NewGodisDB() *GodisDB {
	return &GodisDB{
		data:   NewDict(DictType{HashFunc: StrHash, EqualFunc: StrEqual}),
		expire: NewDict(DictType{HashFunc: StrHash, EqualFunc: StrEqual}),
	}
}

func (db *GodisDB) Lookup(key *Obj) *Obj {
	db.expireIfNeeded(key)
	entry := db.data.Lookup(key)
	if entry != nil {
		return entry.Val
	}
	return nil
}

func (db *GodisDB) expireIfNeeded(key *Obj) {
	entry := db.expire.Lookup(key)
	if entry == nil {
		return
	}

	when := entry.Val.IntVal()
	if when > time.Now().UnixMilli() {
		return
	}

	db.data.Pop(key)
	db.expire.Pop(key)
}

func (db *GodisDB) Set(key, val *Obj) {
	db.data.Insert(key, val)
	db.expire.Pop(key)
}

func (db *GodisDB) Expire(key, val *Obj) {
	db.expire.Insert(key, val)
}

func (db *GodisDB) Cron() {
	keyCount := db.expire.KeyCount()
	cnt := min(100, keyCount)
	for i := int64(0); i < cnt; i++ {
		entry := db.expire.RandomGet()
		if entry.Val.IntVal() < time.Now().Unix() {
			db.data.Pop(entry.Key)
			db.expire.Pop(entry.Key)
		}
	}
}
