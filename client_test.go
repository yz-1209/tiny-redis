package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockIGodisServer struct{}

func (srv *MockIGodisServer) FreeClient(cli *GodisClient)          {}
func (srv *MockIGodisServer) RegisterSendReply(cli *GodisClient)   {}
func (srv *MockIGodisServer) UnRegisterSendReply(cil *GodisClient) {}

func readQuery(cli *GodisClient, query string) {
	for _, b := range []byte(query) {
		cli.queryBuf[cli.queryLen] = b
		cli.queryLen++
	}
}

func TestInlineBuf(t *testing.T) {
	cli := NewGodisClient(0, nil, nil)
	readQuery(cli, "set key val\r\n")
	ok, err := cli.handleInlineBuf()
	assert.Nil(t, err)
	assert.True(t, ok)

	readQuery(cli, "set ")
	ok, err = cli.handleInlineBuf()
	assert.Nil(t, err)
	assert.False(t, ok)

	readQuery(cli, "key ")
	ok, err = cli.handleInlineBuf()
	assert.Nil(t, err)
	assert.False(t, ok)

	readQuery(cli, "val\r\n")
	ok, err = cli.handleInlineBuf()
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, len(cli.args))
}

func TestBulkBuf(t *testing.T) {
	cli := NewGodisClient(0, nil, nil)
	readQuery(cli, "*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n")
	ok, err := cli.handleBulkBuf()
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, 3, len(cli.args))

	readQuery(cli, "*3\r")
	ok, err = cli.handleBulkBuf()
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	readQuery(cli, "\n$3\r\nset\r\n$3")
	ok, err = cli.handleBulkBuf()
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	readQuery(cli, "\r\nkey\r")
	ok, err = cli.handleBulkBuf()
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	readQuery(cli, "\n$3\r\nval\r\n")
	ok, err = cli.handleBulkBuf()
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, 3, len(cli.args))
}

func TestProcessQuery(t *testing.T) {
	db := NewGodisDB()
	srv := &MockIGodisServer{}
	cli := NewGodisClient(0, db, srv)
	readQuery(cli, "*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$3\r\nval\r\n")
	err := cli.ProcessQuery()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cli.args))
	key := NewObject(String, "key")
	val := db.Lookup(key)
	assert.Equal(t, "val", val.StrVal())
    assert.Equal(t, 1, cli.reply.length)
    assert.Equal(t, ReplyOK, cli.reply.First().Val.StrVal())

	readQuery(cli, "set key val2\r\n")
	err = cli.ProcessQuery()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cli.args))
	val2 := db.Lookup(key)
	assert.Equal(t, "val2", val2.StrVal())
    assert.Equal(t, 2, cli.reply.length)
    assert.Equal(t, ReplyOK, cli.reply.Last().Val.StrVal())
}
