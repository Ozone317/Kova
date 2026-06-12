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

// TODO: Make the eviction strategy configurable
// TODO: Support multiple eviction strategies
func evict() {
	switch config.EVICTION_POLICY {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	}
}
