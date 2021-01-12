package consistenthash

import "hash/crc32"

type HashFunc func(data []byte) uint32

type Map struct {
	hashFunc HashFunc
	replicas int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, fn HashFunc) *Map {
	m := &Map{
		replicas: replicas,
		hashFunc: fn,
		hashMap:  make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}
