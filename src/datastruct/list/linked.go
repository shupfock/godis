package list

type node struct {
	val  interface{}
	prev *node
	next *node
}

type LinkedList struct {
	first *node
	last  *node
	size  int
}

func (l *LinkedList) Add(val interface{}) {
	if l == nil {
		panic("list is nil")
	}
	n := &node{
		val: val,
	}
	if l.last == nil {
		l.first = n
		l.last = n
	} else {
		n.prev = l.last
		l.last.next = n
		l.last = n
	}
	l.size++
}

func (l *LinkedList) ForEach(consumer func(int, interface{}) bool) {
	if l == nil {
		panic("list is nil")
	}
	n := l.first
	i := 0
	for n != nil {
		goNext := consumer(i, n.val)
		if !goNext || n.next == nil {
			break
		} else {
			i++
			n = n.next
		}
	}
}
