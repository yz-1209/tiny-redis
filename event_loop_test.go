package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func WriteProc(lp *EventLoop, fd int, arg any) {
	buf := arg.([]byte)
	_, err := Write(fd, buf)
	if err != nil {
		return
	}
	lp.RemoveFileEvent(fd, FE_WRITABLE)
}

func ReadProc(lp *EventLoop, fd int, _ any) {
	buf := make([]byte, 10)
	_, err := Read(fd, buf)
	if err != nil {
		return
	}
	lp.AddFileEvent(fd, FE_WRITABLE, WriteProc, buf)
}

func AcceptProc(lp *EventLoop, fd int, _ any) {
	cfd, err := Accept(fd)
	if err != nil {
		return
	}

	lp.AddFileEvent(cfd, FE_READABLE, ReadProc, nil)
}

func OnceProc(lp *EventLoop, id int, arg any) {
    t := arg.(*testing.T)
    assert.Equal(t, 1, id)
}

func NormalProc(lp *EventLoop, id int, arg any) {
    ch := arg.(chan struct{})
    ch <- struct{}{}
}  

func TestEventLoop(t *testing.T) {
	loop, err := NewEventLoop()
	assert.Nil(t, err)

	sfd, err := TcpServer(6666)
	assert.Nil(t, err)

	loop.AddFileEvent(sfd, FE_READABLE, AcceptProc, nil)
	go loop.Run()

	host := [4]byte{0, 0, 0, 0}
	cfd, err := Connect(host, 6666)
	assert.Nil(t, err)
	msg := "helloworld"
	n, err := Write(cfd, []byte(msg))
	assert.Nil(t, err)
	assert.Equal(t, 10, n)

	buf := make([]byte, 10)
	n, err = Read(cfd, buf)
	assert.Nil(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, msg, string(buf))

    loop.AddTimeEvent(TE_ONCE, 10, OnceProc, t)
    end := make(chan struct{}, 2)
    loop.AddTimeEvent(TE_PERIODIC, 10, NormalProc, end)
    <- end
    <- end
    loop.stop = true
}
