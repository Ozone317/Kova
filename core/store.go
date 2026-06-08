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

	if obj.expiresAtMS > time.Now().UnixMilli() {
		return obj, true
	}

	return nil, false
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
