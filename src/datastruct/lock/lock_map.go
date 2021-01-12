package lock

import (
	"sort"
	"sync"
)

const prime32 = uint32(16777619)

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

type Locks struct {
	table []*sync.RWMutex
}

func New(tableSize int) *Locks {
	table := make([]*sync.RWMutex, tableSize)
	for i := 0; i < tableSize; i++ {
		table[i] = &sync.RWMutex{}
	}
	return &Locks{
		table: table,
	}
}

func (l *Locks) spread(hashCode uint32) uint32 {
	if l == nil {
		panic("lock is nil")
	}
	tableSize := uint32(len(l.table))
	return (tableSize - 1) & uint32(hashCode)
}

func (l *Locks) Lock(key string) {
	index := l.spread(fnv32(key))
	l.table[index].Lock()
}

func (l *Locks) RLock(key string) {
	index := l.spread(fnv32(key))
	l.table[index].RLock()
}

func (l *Locks) Unlock(key string) {
	index := l.spread(fnv32(key))
	l.table[index].Unlock()
}

func (l *Locks) RUnlock(key string) {
	index := l.spread(fnv32(key))
	l.table[index].RUnlock()
}

func (l *Locks) toLockIndices(keys []string, reverse bool) []uint32 {
	indexMap := make(map[uint32]bool)
	for _, key := range keys {
		index := l.spread(fnv32(key))
		indexMap[index] = true
	}
	indices := make([]uint32, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return indices[i] < indices[j]
		} else {
			return indices[i] > indices[j]
		}
	})
	return indices
}

func (l *Locks) Locks(keys ...string) {
	indices := l.toLockIndices(keys, false)
	for _, index := range indices {
		l.table[index].Lock()
	}
}

func (l *Locks) RLocks(keys ...string) {
	indices := l.toLockIndices(keys, false)
	for _, index := range indices {
		l.table[index].RLock()
	}

}

func (l *Locks) Unlocks(keys ...string) {
	indices := l.toLockIndices(keys, false)
	for _, index := range indices {
		l.table[index].Unlock()
	}
}

func (l *Locks) RUnlocks(keys ...string) {
	indices := l.toLockIndices(keys, false)
	for _, index := range indices {
		l.table[index].RUnlock()
	}
}
