package core

import (
	"time"

	"github.com/Ozone317/Kova/config"
)

func evictAllkeysRandom() {
	numToEvict := int(config.MAX_KEYS * config.ALL_KEYS_EVICTION_RATIO)

	for k := range store {
		delete(store, k)
		numToEvict--
		if numToEvict == 0 {
			break
		}
	}

}

func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF // shaving off the first 8 bits to keep the time in 24 bits as Redis does
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()

	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}

	return (0x00FFFFFF - lastAccessedAt) + c
}

func populateEvictionPool() {
	sampleSize := 5
	for k := range store {
		// push the key in the pool
		// reduce sample size
		// break if sample size reaches 0
		ePool.Push(k, store[k].lastAccessedTimeMS)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}

func evictAllkeysLRU() {
	populateEvictionPool()
	evictCount := int16(config.ALL_KEYS_EVICTION_RATIO * float64(config.MAX_KEYS))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		Del([]string{item.key})
	}
}

// TODO: Make the eviction strategy configurable
// TODO: Support multiple eviction strategies
func evict() {
	switch config.EVICTION_POLICY {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	case "allkeys-lru":
		evictAllkeysLRU()
	}
}
