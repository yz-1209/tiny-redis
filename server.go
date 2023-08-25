package main

import (
	"log"

	"golang.org/x/exp/constraints"
)

type GodisServer struct {
	fd             int
	port           int
	maxClientLimit int
	lp             *EventLoop
	db             *GodisDB
	clients        map[int]*GodisClient
}

func NewGodisServer(port, maxClientLimit int) *GodisServer {
	return &GodisServer{
		port:           port,
		maxClientLimit: maxClientLimit,
		db:             NewGodisDB(),
		clients:        make(map[int]*GodisClient),
	}
}

func (srv *GodisServer) Run() (err error) {
	srv.fd, err = TcpServer(srv.port)
	if err != nil {
		return err
	}

	srv.lp, err = NewEventLoop()
	if err != nil {
		return err
	}

	srv.lp.AddFileEvent(srv.fd, FE_READABLE, srv.AcceptHandler, nil)
	srv.lp.AddTimeEvent(TE_PERIODIC, 100, srv.Cron, nil)
	srv.lp.Run()
	return
}

func (srv *GodisServer) AcceptHandler(lp *EventLoop, fd int, _ any) {
	cfd, err := Accept(fd)
	if err != nil {
		log.Println("accept failed: ", err)
		return
	}

	if len(srv.clients) >= srv.maxClientLimit {
		log.Println("exceed max client limit, close conn...")
		Close(fd)
		return
	}

	cli := NewGodisClient(cfd, srv.db, srv)
	srv.clients[cfd] = cli
	srv.lp.AddFileEvent(cfd, FE_READABLE, cli.ReadQuery, nil)
}

func (srv *GodisServer) Cron(lp *EventLoop, id int, _ any) {
	srv.db.Cron()
}

func (srv *GodisServer) FreeClient(cli *GodisClient) {
	delete(srv.clients, cli.fd)
	srv.lp.RemoveFileEvent(cli.fd, FE_READABLE)
	srv.lp.RemoveFileEvent(cli.fd, FE_WRITABLE)
}

func (srv *GodisServer) RegisterSendReply(cli *GodisClient) {
	srv.lp.AddFileEvent(cli.fd, FE_WRITABLE, cli.SendReply, cli)
}

func (srv *GodisServer) UnRegisterSendReply(cli *GodisClient) {
	srv.lp.RemoveFileEvent(cli.fd, FE_WRITABLE)
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
