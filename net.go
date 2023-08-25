package main

import (
	"fmt"
	"golang.org/x/sys/unix"
)

const Backlog = 64

func TcpServer(port int) (int, error) {
	// create ipv4 tcp socket
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return -1, fmt.Errorf("init socket failed: %v", err)
	}

	// the SO_REUSEPORT option allows multiple sockets on the same host to bind to the same port,
	// and is intended to improve the performance of multithreaded network server applications running on top of multicore systems.
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, port)
	if err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("set SO_REUSEPORT failed: %v", err)
	}

	var addr unix.SockaddrInet4
	addr.Port = port

	// golang will set addr.Addr = any(0)
	// golang will handle htons
	err = unix.Bind(fd, &addr)
	if err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("bind addr failed: %v", err)
	}

	// the Backlog parameter specifies the number of pending connections the queue will hold.
	err = unix.Listen(fd, Backlog)
	if err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("listen socket failed: %v", err)
	}
	return fd, err
}

func Accept(fd int) (int, error) {
	// ignore client addr
	nfd, _, err := unix.Accept(fd)
	return nfd, err
}

func Connect(host [4]byte, port int) (int, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return -1, fmt.Errorf("init socket failed: %v", err)
	}

	var addr unix.SockaddrInet4
	addr.Addr = host
	addr.Port = port
	err = unix.Connect(fd, &addr)
	if err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("connect failed: %v", err)
	}
	return fd, nil
}

func Read(fd int, buf []byte) (int, error) {
	return unix.Read(fd, buf)
}

func Write(fd int, buf []byte) (int, error) {
	return unix.Write(fd, buf)
}

func Close(fd int) {
    unix.Close(fd)
}
