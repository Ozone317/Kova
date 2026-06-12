package core

import (
	"time"

	"github.com/Ozone317/Kova/config"
)

var store map[string]*Object
var expires map[*Object]uint64

func init() {
	store = make(map[string]*Object)
	expires = make(map[*Object]uint64)
}

func setExpiry(obj *Object, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
}

func NewObject(value interface{}, objectType uint8, objectEncoding uint8, durationMS int64) *Object {
	obj := &Object{
		value:              value,
		lastAccessedTimeMS: getCurrentClock(),
		typeEncoding:       (objectType << 4) | objectEncoding,
	}

	if durationMS > 0 {
		setExpiry(obj, durationMS)
	}

	return obj
}

func Put(key string, obj *Object) {
	if len(store) >= config.MAX_KEYS {
		evict()
	}
	obj.lastAccessedTimeMS = getCurrentClock()
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++

	store[key] = obj
}

func Get(key string) (*Object, bool) {
	obj, ok := store[key]
	if !ok {
		return nil, false
	}

	if hasExpired(obj) {
		Del([]string{key})
		return nil, false
	}
	obj.lastAccessedTimeMS = getCurrentClock()
	return obj, true
}

func Ttl(key string) int64 {
	obj, ok := Get(key)
	if !ok {
		return -2
	}

	exp, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return -1
	}

	if exp > uint64(time.Now().UnixMilli()) {
		return int64((exp - uint64(time.Now().UnixMilli())) / 1000)
	}
	Del([]string{key})
	return -2
}

func Del(keys []string) int64 {
	var deleted int64 = 0
	for _, key := range keys {
		obj, ok := store[key]
		if !ok {
			continue
		}

		delete(store, key)
		delete(expires, obj)
		deleted++
		KeyspaceStat[0]["keys"]--
	}

	return deleted
}

func Expire(key string, ttlSeconds int64) int64 {
	obj, ok := store[key]
	if !ok {
		return 0
	}

	setExpiry(obj, ttlSeconds*1000)
	obj.lastAccessedTimeMS = getCurrentClock()
	return 1
}
