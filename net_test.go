package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func EchoServer(port int, ready chan struct{}, logf func(string, ...any)) {
	sfd, err := TcpServer(port)
	if err != nil {
		logf("start tcp server failed: %v", err)
		return
	}
	ready <- struct{}{}
	cfd, err := Accept(sfd)
	if err != nil {
		logf("accept failed: %v", err)
		return
	}

	logf("accept conn: cfd = %v", cfd)
	buf := make([]byte, 10)
	_, err = Read(cfd, buf)
	if err != nil {
		logf("read failed: %v", err)
		return
	}
	_, err = Write(cfd, buf)
	if err != nil {
		logf("write failed: %v", err)
	}
}

func TestEchoServer(t *testing.T) {
	port := 6667
	ready := make(chan struct{})
	go EchoServer(port, ready, t.Logf)

	t.Logf("go run echo server...")
	<-ready
	host := [4]byte{127, 0, 0, 1}
	cfd, err := Connect(host, port)
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
}
