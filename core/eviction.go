package core

func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

// TODO: Make the eviction strategy configurable
// TODO: Support multiple eviction strategies
func evict() {
	evictFirst()
}
