package dict

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
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

type Shard struct {
	m     map[string]interface{}
	mutex sync.RWMutex
}

type ConcurrentDict struct {
	table []*Shard
	count int32
}

func computeCapacity(param int) (size int) {
	if param <= 16 {
		return 16
	}
	n := param - 1
	fmt.Println(n)
	n |= n >> 1
	fmt.Println(n)
	n |= n >> 2
	fmt.Println(n)
	n |= n >> 4
	fmt.Println(n)
	n |= n >> 8
	fmt.Println(n)
	n |= n >> 16
	fmt.Println(n)
	if n < 0 {
		return math.MaxInt32
	}
	return int(n + 1)
}

func (shard *Shard) RandomKey() string {
	if shard == nil {
		panic("shard is nil")
	}
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	for key := range shard.m {
		return key
	}
	return ""
}

func NewConcurrent(ShardCount int) *ConcurrentDict {
	shardCount := computeCapacity(ShardCount)
	table := make([]*Shard, shardCount)
	for i := 0; i < ShardCount; i++ {
		table[i] = &Shard{
			m: make(map[string]interface{}),
		}
	}
	d := &ConcurrentDict{
		count: 0,
		table: table,
	}
	return d
}

func (dict *ConcurrentDict) addCount() {
	if dict == nil {
		panic("dict is nil")
	}
	dict.count++
}

func (dict *ConcurrentDict) spread(hashCode uint32) uint32 {
	if dict == nil {
		panic("dict id nil")
	}
	tableSize := uint32(len(dict.table))
	return (tableSize - 1) & uint32(hashCode)
}

func (dict *ConcurrentDict) getShard(index uint32) *Shard {
	if dict == nil {
		panic("dict is nil")
	}
	return dict.table[index]
}

func (dict *ConcurrentDict) Get(key string) (val interface{}, exists bool) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := fnv32(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.RLock()
	val, exists = shard.m[key]
	shard.mutex.RUnlock()
	return
}

func (dict *ConcurrentDict) Len() int {
	if dict == nil {
		panic("dict is nil")
	}
	return int(atomic.LoadInt32(&dict.count))
}

func (dict *ConcurrentDict) Put(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := fnv32(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.m[key]; ok {
		shard.m[key] = val
		return 0
	}
	shard.m[key] = val
	dict.addCount()
	return 1
}

func (dict *ConcurrentDict) PutIfAbsent(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := fnv32(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock(0)

	if _, ok := shard.m[key]; ok {
		return 0
	} else {
		shard.m[key] = val
		dict.addCount()
		return 1
	}
}

func (dict *ConcurrentDict) PutIfExists(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := fnv32(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.m[key]; ok {
		shard.m[key] = val
		return 1
	}
	return 0
}

func (dict *ConcurrentDict) Remove(key string) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := fnv32(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.m[key]; ok {
		delete(shard.m, key)
		return 1
	}
	return 0
}

func (dict *ConcurrentDict) ForEach(consumer Consumer) {
	if dict == nil {
		panic("dict is nil")
	}

	for _, shard := range dict.table {
		for key, value := range shard.m {
			shard.mutex.RLock()
			continues := consumer(key, value)
			shard.mutex.RUnlock()
			if continues {
				return
			}
		}
	}
}

func (dict *ConcurrentDict) Keys() []string {
	keys := make([]string, dict.Len())
	i := 0
	dict.ForEach(func(key string, val interface{}) bool {
		if i < len(keys) {
			keys[i] = key
			i++
		} else {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

func (dict *ConcurrentDict) RandomKeys(limit int) []string {
	if dict == nil {
		panic("dict is nil")
	}
	size := dict.Len()
	if limit >= size {
		return dict.Keys()
	}
	shardCount := len(dict.table)

	result := make([]string, limit)
	for i := 0; i < limit; {
		shard := dict.getShard(uint32(rand.Intn(shardCount)))
		if shard == nil {
			continue
		}
		key := shard.RandomKey()
		if key != "" {
			result[i] = key
			i++
		}
	}
	return result
}

func (dict *ConcurrentDict) RandomDistinctKeys(limit int) []string {
	size := dict.Len()
	if limit >= size {
		return dict.Keys()
	}

	shardCount := len(dict.table)
	result := make(map[string]bool)
	for len(result) < limit {
		shardIndex := uint32(rand.Intn(shardCount))
		shard := dict.getShard(shardIndex)
		if shard == nil {
			continue
		}
		key := shard.RandomKey()
		if key != "" {
			result[key] = true
		}
	}
	arr := make([]string, limit)
	i := 0
	for k := range result {
		arr[i] = k
	}
	return arr
}
