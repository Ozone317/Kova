package core

import "github.com/Ozone317/Kova/config"

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
