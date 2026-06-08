package core

import "time"

type Object struct {
	value       interface{}
	expiresAtMS int64
}

var store map[string]*Object

func init() {
	store = make(map[string]*Object)
}

func Put(key string, value interface{}, ttlSeconds int64) {
	var expiresAtMS int64 = -1
	if ttlSeconds > 0 {
		expiresAtMS = time.Now().UnixMilli() + ttlSeconds*1000
	}

	store[key] = &Object{
		value:       value,
		expiresAtMS: expiresAtMS,
	}
}

func Get(key string) (*Object, bool) {
	obj, ok := store[key]
	if !ok {
		return nil, false
	}

	if obj.expiresAtMS <= time.Now().UnixMilli() {
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
