package main

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	object	[]byte
	next	*node
}

type queue struct {
	dummy	*node
	tail	*node
}

func newQueue() *queue {
	q := new(queue)
	q.dummy = new(node)
	q.tail = q.dummy
	return q
}

func (q *queue) enqueue(object []byte) {
	newNode := new(node)
	newNode.object = object

	added := false

	var oldTail *node
	for !added {
		oldTail = q.tail
		oldTailNext := oldTail.next

		if q.tail != oldTail {
			continue
		}

		if oldTailNext != nil {
			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(oldTail), unsafe.Pointer(oldTailNext))
			continue
		}

		added = atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&oldTail.next)), unsafe.Pointer(oldTailNext), unsafe.Pointer(newNode))
	}

	atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(oldTail), unsafe.Pointer(newNode))
}

func (q *queue) dequeue() ([]byte, bool) {
	var object []byte
	removed := false

	for !removed {
		oldDummy := q.dummy
		oldHead := oldDummy.next
		oldTail := q.tail

		if q.dummy != oldDummy {
			continue
		}

		if oldHead == nil {
			return nil, false
		}

		if oldTail == oldDummy {
			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(oldTail), unsafe.Pointer(oldHead))
			continue
		}
		object = oldHead.object
		removed = atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.dummy)), unsafe.Pointer(oldDummy), unsafe.Pointer(oldHead))
	}
	return object, true
}
