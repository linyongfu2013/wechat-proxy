package wechat

import (
	"errors"
	"sync"
	"time"
)

type cacheItem struct {
	value  interface{}
	expire int64
}

type CacheMap struct {
	m      map[string]cacheItem
	lock   sync.RWMutex
	config struct {
		duration time.Duration // duration for cache item in memory.
		limit    int           // limit for cache item count.
	}
}

func NewCacheMap(duration time.Duration, limit int) *CacheMap {
	tm := new(CacheMap)
	tm.m = make(map[string]cacheItem)
	tm.config.duration = duration
	tm.config.limit = limit
	return tm
}

func (tm *CacheMap) Set(key string, value interface{}) {
	expire := time.Now().Add(tm.config.duration).Unix()
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.m[key] = cacheItem{value: value, expire: expire}
}

func (tm *CacheMap) Get(key string) (value interface{}, success bool) {
	tm.lock.RLock()
	defer tm.lock.RUnlock()
	if token, ok := tm.m[key]; ok {
		value = token.value
		success = (time.Now().Unix() < token.expire)
	}
	return
}

func (tm *CacheMap) Remove(key string) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	delete(tm.m, key)
}

func (tm *CacheMap) Shrink() {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	if len(tm.m) < tm.config.limit {
		return
	}
	now := time.Now().Unix()
	for k := range tm.m {
		if tm.m[k].expire < now {
			delete(tm.m, k)
		}
	}
}

var ErrCacheTimeout = errors.New("cache timeout")
