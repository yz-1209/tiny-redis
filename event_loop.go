package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/unix"
)

type FeType uint8

const (
	FE_READABLE FeType = iota
	FE_WRITABLE
)

func (ft FeType) ToEpollEvent() uint32 {
	if ft == FE_READABLE {
		return unix.EPOLLIN
	}

	return unix.EPOLLOUT
}

type TeType uint8

const (
	TE_PERIODIC TeType = iota
	TE_ONCE
)

type FileProc func(lp *EventLoop, fd int, arg any)
type TimeProc func(lp *EventLoop, id int, arg any)

type FileEvent struct {
	fd   int
	mask FeType
	proc FileProc
	arg  any
	next *FileEvent
}

type TimeEvent struct {
	id       int
	mask     TeType
	when     int64 // ms
	interval int64 // ms
	proc     TimeProc
	arg      any
	next     *TimeEvent
}

type EventLoop struct {
	FileEvents *FileEvent
	TimeEvents *TimeEvent
	fd         int
	nextId     int
	stop       bool
}

func NewEventLoop() (*EventLoop, error) {
	epollFd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("create epoll fd failed: %v", err)
	}

	return &EventLoop{
		fd:     epollFd,
		nextId: 1,
	}, nil
}

func (lp *EventLoop) nearestTime() int64 {
	nearest := time.Now().UnixMilli() + 1000
	p := lp.TimeEvents
	for p != nil {
		if p.when != 0 && p.when < nearest {
			nearest = p.when
		}
		p = p.next
	}
	return nearest
}

func (lp *EventLoop) searchFileEvent(fd int, mask FeType) *FileEvent {
	p := lp.FileEvents
	for p != nil {
		if p.fd == fd && p.mask == mask {
			return p
		}
		p = p.next
	}
	return nil
}

func (lp *EventLoop) getRegisteredEpollEvent(fd int) uint32 {
	var epollEvent uint32
	if lp.searchFileEvent(fd, FE_READABLE) != nil {
		epollEvent |= unix.EPOLLIN
	}
	if lp.searchFileEvent(fd, FE_WRITABLE) != nil {
		epollEvent |= unix.EPOLLOUT
	}
	return epollEvent
}

func (lp *EventLoop) AddFileEvent(fd int, mask FeType, proc FileProc, arg any) {
	epollEvent := lp.getRegisteredEpollEvent(fd)
	if epollEvent&mask.ToEpollEvent() != 0 {
		return
	}

	op := 0
	if epollEvent == 0 {
		op = unix.EPOLL_CTL_ADD
	} else {
		op = unix.EPOLL_CTL_MOD
	}

	epollEvent |= mask.ToEpollEvent()
	err := unix.EpollCtl(lp.fd, op, fd, &unix.EpollEvent{Fd: int32(fd), Events: epollEvent})
	if err != nil {
		log.Println("epoll ctl failed: ", err)
		return
	}

	fe := &FileEvent{
		fd:   fd,
		mask: mask,
		proc: proc,
		arg:  arg,
		next: lp.FileEvents,
	}
	lp.FileEvents = fe
	log.Printf("add file event, fd = %v, mask = %v", fd, mask)
}

func (lp *EventLoop) RemoveFileEvent(fd int, mask FeType) {
	op := unix.EPOLL_CTL_DEL
	epollEvent := lp.getRegisteredEpollEvent(fd)
	epollEvent &= ^mask.ToEpollEvent()
	if epollEvent != 0 {
		op = unix.EPOLL_CTL_MOD
	}

	err := unix.EpollCtl(lp.fd, op, fd, &unix.EpollEvent{Fd: int32(fd), Events: epollEvent})
	if err != nil {
		log.Println("epoll del failed:", err)
	}

	var prev, curr *FileEvent = nil, lp.FileEvents
	for curr != nil && (curr.fd != fd || curr.mask != mask) {
		prev, curr = curr, curr.next
	}

	if curr != nil {
		if prev == nil {
			lp.FileEvents = curr.next
		} else {
			prev.next = curr.next
		}
	}
}

func (lp *EventLoop) AddTimeEvent(mask TeType, interval int64, proc TimeProc, extra any) int {
	te := &TimeEvent{
		id:       lp.nextId,
		mask:     mask,
		when:     time.Now().UnixMilli() + interval,
		interval: interval,
		proc:     proc,
		arg:      extra,
		next:     lp.TimeEvents,
	}
	lp.TimeEvents = te
	lp.nextId++
	log.Printf("add time event, id = %v, mask = %v\n", te.id, mask)
	return te.id
}

func (lp *EventLoop) RemoveTimeEvent(id int) {
	curr := lp.TimeEvents
	var prev *TimeEvent
	for curr != nil && curr.id != id {
		prev, curr = curr, curr.next
	}

	if curr != nil {
		if prev == nil {
			lp.TimeEvents = curr.next
		} else {
			prev.next = curr.next
		}
	}
}

func (lp *EventLoop) WaitEvents() (fileEvents []*FileEvent, timeEvents []*TimeEvent) {
	timeout := lp.nearestTime() - time.Now().UnixMilli()
	if timeout <= 0 {
		timeout = 10
	}

	var events [128]unix.EpollEvent
	// log.Printf("start to epoll wait, timeout = %v\n", timeout)
	n, err := unix.EpollWait(lp.fd, events[:], int(timeout))
	if err != nil {
		log.Printf("epoll wait warnning: %v\n", err)
	}

	for i := 0; i < n; i++ {
		if events[i].Events&unix.EPOLLIN != 0 {
			fe := lp.searchFileEvent(int(events[i].Fd), FE_READABLE)
			if fe != nil {
				fileEvents = append(fileEvents, fe)
			}
		}
		if events[i].Events&unix.EPOLLOUT != 0 {
			fe := lp.searchFileEvent(int(events[i].Fd), FE_WRITABLE)
			if fe != nil {
				fileEvents = append(fileEvents, fe)
			}
		}
	}

	now := time.Now().UnixMilli()
	p := lp.TimeEvents
	for p != nil {
		if p.when < now {
			timeEvents = append(timeEvents, p)
		}
		p = p.next
	}

	// log.Printf("finished collect events, file events = %v, time events = %v", len(fileEvents), len(timeEvents))
	return
}

func (lp *EventLoop) ProcessEvents(fileEvents []*FileEvent, timeEvents []*TimeEvent) {
	for _, event := range timeEvents {
		event.proc(lp, event.id, event.arg)
		if event.mask == TE_ONCE {
			lp.RemoveTimeEvent(event.id)
		} else {
			event.when = time.Now().UnixMilli() + event.interval
		}
	}

	for _, event := range fileEvents {
		event.proc(lp, event.fd, event.arg)
	}
}

func (lp *EventLoop) Run() {
	for !lp.stop {
		fileEvents, timeEvents := lp.WaitEvents()
		lp.ProcessEvents(fileEvents, timeEvents)
	}
}
