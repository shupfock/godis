package db

import (
	"godis/src/datastruct/dict"
	List "godis/src/datastruct/list"
	"godis/src/datastruct/lock"
	"sync"
	"time"
)

type DataEntity struct {
	Data interface{}
}

const (
	dataDictSize = 1 << 16
	ttlDictSize  = 1 << 10
	lockerSize   = 128
	aofQueueSize = 1 << 16
)

type DB struct {
	Data dict.Dict

	TTLMap dict.Dict

	sunMap dict.Dict

	Locker *lock.Locks

	interval time.Duration

	stopWorld sync.WaitGroup
}

func NewDB() *DB {
	db := &DB{
		Data:     dict.NewConcurrent(dataDictSize),
		TTLMap:   dict.NewConcurrent(ttlDictSize),
		Locker:   lock.New(lockerSize),
		interval: 5 * time.Second,
	}
}

func (db *DB) IsExpired(key string) bool {
	rawExpireTime, ok := db.TTLMap.Get(key)
	if !ok {
		return false
	}
	expireTime, _ := rawExpireTime.(time.Time)
	expired := time.Now().After(expireTime)
	if expired {
		db.Remove(key)
	}
	return expired
}

func (db *DB) Get(key string) (*DataEntity, bool) {
	db.stopWorld.Wait()

	raw, ok := db.Data.Get(key)
	if !ok {
		return nil, false
	}
	if db.IsExpired(key) {
		return nil, false
	}
	entity, _ := raw.(*DataEntity)
	return entity, true
}

func (db *DB) Remove(key string) {
	db.stopWorld.Wait()
	db.Data.Remove(key)
	db.TTLMap.Remove(key)
}

func (db *DB) CleanExpired() {
	now := time.Now()
	toRemove := &List.LinkedList{}
	db.TTLMap.ForEach(func(key string, val interface{}) bool {
		expiredTime, _ := val.(time.Time)
		if now.After(expiredTime) {
			db.Data.Remove(key)
			toRemove.Add(key)
		}
		return true
	})
	toRemove.ForEach(func(i int, val interface{}) bool {
		key, _ := val.(string)
		db.TTLMap.Remove(key)
		return true
	})
}

func (db *DB) TimerTash() {
	ticker := time.NewTicker(db.interval)
	go func() {
		for range ticker.C {
			db.CleanExpired()
		}
	}()
}
