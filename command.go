package main

import (
	"fmt"
	"log"
	"time"
)

const (
	GodisCmdGet    = "get"
	GodisCmdSet    = "set"
	GodisCmdExpire = "expire"
	GodisCmdQuit   = "quit"

	ReplyWrongType         = "-ERR: wrong type\r\n"
	ReplyMinusOne          = "-1\r\n"
	ReplyOK                = "+OK\r\n"
	ReplyUnknownCmd        = "-ERR: unknow command\r\n"
	ReplyWrongNumberOfArgs = "-ERR: wrong number of args\r\n"
)

var CmdTable = map[string]*GodisCommand{
	GodisCmdGet:    &GodisCommand{GodisCmdGet, getCmd, 2},
	GodisCmdSet:    &GodisCommand{GodisCmdSet, setCmd, 3},
	GodisCmdExpire: &GodisCommand{GodisCmdExpire, expireCmd, 3},
}

type GodisCommand struct {
	name  string
	proc  func(args []*Obj, db *GodisDB) string
	arity int // the number of arguments
}

func getCmd(args []*Obj, db *GodisDB) string {
	key := args[1]
	val := db.Lookup(key)
	if val == nil {
		return ReplyMinusOne
	}
	if val.Type != String {
		return ReplyWrongType
	}
	valStr := val.StrVal()
	return fmt.Sprintf("$%d%v\r\n", len(valStr), valStr)
}

func setCmd(args []*Obj, db *GodisDB) string {
	key, val := args[1], args[2]
	if val.Type != String {
		return ReplyWrongType
	}
	db.Set(key, val)
	return ReplyOK
}

func expireCmd(args []*Obj, db *GodisDB) string {
	key, val := args[1], args[2]
	if val.Type != String {
		return ReplyWrongType
	}

	expire := time.Now().UnixMilli() + val.IntVal()*1000
	expObj := NewObjectInt(expire)
	db.Expire(key, expObj)
	expObj.DecrRefCount()
	return ReplyOK
}

func processCmd(args []*Obj, db *GodisDB) string {
	cmdStr := args[0].StrVal()
	log.Printf("process command: cmd = %v", cmdStr)
	var reply string
	switch cmd := CmdTable[cmdStr]; {
	case cmd == nil:
		reply = ReplyUnknownCmd
	case cmd.arity != len(args):
		reply = ReplyWrongNumberOfArgs
	default:
		reply = cmd.proc(args, db)
	}
	return reply
}
