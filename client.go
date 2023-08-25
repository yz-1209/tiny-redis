package main

import (
	"bytes"
	"errors"
	"log"
	"strconv"
)

type CmdType byte

const (
	CmdUnknown CmdType = iota
	CmdInline
	CmdBulk
)

const (
	GodisMaxBulk   int = 1024 * 4
	GodisMaxInline int = 1024 * 4
	GodisIOBuffer  int = 1024 * 16
)

var (
	ErrTooBigInlineCmd     = errors.New("too big inline cmd")
	ErrTooBigBulkCmd       = errors.New("too big bulk cmd")
	ErrUnknownGodisCmdType = errors.New("unknown godis command type")
	ErrExpectedBulkLength  = errors.New("expect $ for bulk length")
	ErrExpectedBulkEnd     = errors.New("expect CRLF for bulk end")
)

type IGodisServer interface {
	FreeClient(cli *GodisClient)
	RegisterSendReply(cli *GodisClient)
	UnRegisterSendReply(cli *GodisClient)
}

type GodisClient struct {
	fd       int
	bulkLen  int
	bulkNum  int
	sentLen  int
	queryLen int
	queryBuf []byte
	cmdType  CmdType
	args     []*Obj
	reply    *List
	db       *GodisDB
	srv      IGodisServer
}

func NewGodisClient(fd int, db *GodisDB, srv IGodisServer) *GodisClient {
	return &GodisClient{
		fd:       fd,
		db:       db,
		srv:      srv,
		queryBuf: make([]byte, GodisIOBuffer),
		reply:    NewList(ListType{StrEqual}),
	}
}

func (cli *GodisClient) ReadQuery(lp *EventLoop, fd int, _ any) {
	if len(cli.queryBuf)-cli.queryLen < GodisMaxBulk {
		cli.queryBuf = append(cli.queryBuf, make([]byte, GodisMaxBulk)...)
	}

	n, err := Read(cli.fd, cli.queryBuf[cli.queryLen:])
	if err != nil {
		log.Printf("cli %v read failed: %v\n", fd, err)
		cli.free()
		return
	}

	cli.queryLen += n
	err = cli.ProcessQuery()
	if err != nil {
		cli.free()
	}
}

func (cli *GodisClient) ProcessQuery() (err error) {
	for cli.queryLen > 0 {
		if cli.cmdType == CmdUnknown {
			if cli.queryBuf[0] == '*' {
				cli.cmdType = CmdBulk
			} else {
				cli.cmdType = CmdInline
			}
		}

		ok := false
		if cli.cmdType == CmdInline {
			ok, err = cli.handleInlineBuf()
		} else {
			ok, err = cli.handleBulkBuf()
		}

		if err != nil || !ok {
			return
		}

		if len(cli.args) > 0 {
			// handle "quit" special command
			if cli.args[0].StrVal() == GodisCmdQuit {
				cli.free()
				return
			}

			reply := processCmd(cli.args, cli.db)
			cli.reply.Append(NewObject(String, reply))
			cli.srv.RegisterSendReply(cli)
		}
		cli.reset()
	}
	return nil
}

func (cli *GodisClient) handleInlineBuf() (bool, error) {
	idx := bytes.Index(cli.queryBuf[:cli.queryLen], []byte("\r\n"))
	if idx < 0 {
		if cli.queryLen > GodisMaxInline {
			return false, ErrTooBigInlineCmd
		} else {
			return false, nil
		}
	}

	parts := bytes.Split(cli.queryBuf[:idx], []byte(" "))
	cli.queryBuf = cli.queryBuf[idx+2:]
	cli.queryLen -= idx + 2
	cli.args = make([]*Obj, len(parts))
	for i, part := range parts {
		cli.args[i] = NewObject(String, string(part))
	}
	return true, nil
}

func (cli *GodisClient) handleBulkBuf() (bool, error) {
	if cli.bulkNum == 0 {
		idx, err := cli.findLineInQuery()
		if idx < 0 || err != nil {
			return false, err
		}

		num, err := cli.getNumInQuery(1, idx)
		if err != nil {
			return false, err
		}
		if num == 0 {
			return true, nil
		}

		cli.bulkNum = num
		cli.args = make([]*Obj, num)
	}

	for cli.bulkNum > 0 {
		if cli.bulkLen == 0 {
			idx, err := cli.findLineInQuery()
			if idx < 0 {
				return false, err
			}

			if cli.queryBuf[0] != '$' {
				return false, ErrExpectedBulkLength
			}

			blen, err := cli.getNumInQuery(1, idx)
			if err != nil || blen == 0 {
				return false, err
			}

			if blen > GodisMaxBulk {
				return false, ErrTooBigBulkCmd
			}
			cli.bulkLen = blen
		}
		if cli.queryLen < cli.bulkLen+2 {
			return false, nil
		}
		idx := cli.bulkLen
		if cli.queryBuf[idx] != '\r' || cli.queryBuf[idx+1] != '\n' {
			return false, ErrExpectedBulkEnd
		}
		cli.args[len(cli.args)-cli.bulkNum] = NewObject(String, string(cli.queryBuf[:idx]))
		cli.queryBuf = cli.queryBuf[idx+2:]
		cli.queryLen -= idx + 2
		cli.bulkLen = 0
		cli.bulkNum--
	}
	return true, nil
}

func (cli *GodisClient) findLineInQuery() (int, error) {
	idx := bytes.Index(cli.queryBuf[:cli.queryLen], []byte("\r\n"))
	if idx < 0 && cli.queryLen > GodisMaxInline {
		return idx, ErrTooBigInlineCmd
	}
	return idx, nil
}

func (cli *GodisClient) getNumInQuery(start, end int) (int, error) {
	num, err := strconv.Atoi(string(cli.queryBuf[start:end]))
	cli.queryBuf = cli.queryBuf[end+2:]
	cli.queryLen -= end + 2
	return num, err
}

func (cli *GodisClient) SendReply(lp *EventLoop, fd int, _ any) {
	for cli.reply.length > 0 {
		first := cli.reply.First()
		buf := []byte(first.Val.StrVal())
		bufLen := len(buf)
		if cli.sentLen < bufLen {
			n, err := Write(fd, buf[cli.sentLen:])
			if err != nil {
				log.Printf("send reply failed: %v\n", err)
				cli.free()
				return
			}

			cli.sentLen += n
			if cli.sentLen < bufLen {
				return
			}

			cli.reply.DelNode(first)
			first.Val.DecrRefCount()
			cli.sentLen = 0
		}
	}

	if cli.reply.length == 0 {
		cli.sentLen = 0
		cli.srv.UnRegisterSendReply(cli)
	}
}

func (cli *GodisClient) reset() {
	cli.freeArgs()
	cli.cmdType = CmdUnknown
	cli.bulkLen = 0
	cli.bulkNum = 0
}

func (cli *GodisClient) free() {
	cli.freeArgs()
	cli.freeReplyList()
	cli.srv.FreeClient(cli)
	Close(cli.fd)
}

func (cli *GodisClient) freeReplyList() {
	for cli.reply.length != 0 {
		n := cli.reply.First()
		cli.reply.DelNode(n)
		n.Val.DecrRefCount()
	}
}

func (cli *GodisClient) freeArgs() {
	for _, arg := range cli.args {
		arg.DecrRefCount()
	}
}
