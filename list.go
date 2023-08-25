package main

type ListNode struct {
	Val  *Obj
	prev *ListNode
	next *ListNode
}

type ListType struct {
	EqualFunc func(a, b *Obj) bool
}

type List struct {
	ListType

	head   *ListNode
	tail   *ListNode
	length int
}

func NewList(listType ListType) *List {
	head, tail := &ListNode{}, &ListNode{}
	head.next = tail
	tail.prev = head
	return &List{
		ListType: listType,
		head:     head,
		tail:     tail,
		length:   0,
	}
}

func (l *List) First() *ListNode {
	if l.length == 0 {
		return nil
	}

	return l.head.next
}

func (l *List) Last() *ListNode {
	if l.length == 0 {
		return nil
	}

	return l.tail.prev
}
func (l *List) Find(val *Obj) *ListNode {
	p := l.head.next
	for p != l.tail {
		if l.EqualFunc(p.Val, val) {
			return p
		}
		p = p.next
	}
	return nil
}

func (l *List) LPush(val *Obj) {
	n := &ListNode{Val: val, prev: l.head, next: l.head.next}
	l.head.next.prev = n
	l.head.next = n
	l.length++
}

func (l *List) Append(val *Obj) {
	n := &ListNode{Val: val, prev: l.tail.prev, next: l.tail}
	l.tail.prev.next = n
	l.tail.prev = n
	l.length++
}

func (l *List) DelNode(n *ListNode) {
	n.prev.next = n.next
	n.next.prev = n.prev
	l.length--
}

func (l *List) Delete(val *Obj) {
	n := l.Find(val)
	if n != nil {
		l.DelNode(n)
	}
}
