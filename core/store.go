package core

import (
	"time"

	"github.com/Ozone317/Kova/config"
)

var store map[string]*Object

func init() {
	store = make(map[string]*Object)
}

func NewObject(value interface{}, objectType uint8, objectEncoding uint8, durationMS int64) *Object {
	var expiresAtMS int64 = -1
	if durationMS > 0 {
		expiresAtMS = time.Now().UnixMilli() + durationMS
	}

	return &Object{
		value:        value,
		expiresAtMS:  expiresAtMS,
		typeEncoding: (objectType << 4) | objectEncoding,
	}
}

func Put(key string, obj *Object) {
	if len(store) >= config.MAX_KEYS {
		evict()
	}

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

	if obj.expiresAtMS != -1 && obj.expiresAtMS <= time.Now().UnixMilli() {
		delete(store, key)
		return nil, false
	}

	return obj, true
}

func Ttl(key string) int64 {
	obj, ok := Get(key)
	if !ok {
		return -2
	}

	if obj.expiresAtMS == -1 {
		return -1
	}

	if obj.expiresAtMS > time.Now().UnixMilli() {
		return int64((obj.expiresAtMS - time.Now().UnixMilli()) / 1000)
	}
	return -2
}

func Del(keys []string) int64 {
	var deleted int64 = 0
	for _, key := range keys {
		_, ok := store[key]
		if !ok {
			continue
		}

		delete(store, key)
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

	obj.expiresAtMS = time.Now().UnixMilli() + ttlSeconds*1000
	return 1
}
