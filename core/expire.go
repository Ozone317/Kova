// On a fixed timer (every 1s), we run a background cycle that proactively
// hunts for and deletes expired keys. We cannot afford to scan the entire
// keyspace every 1s - that would block the server. Instead we use a
// probabilistic approach:
//
//   1. Sample 20 keys at random from the set of keys that have a TTL set.
//      We only sample from volatile keys (keys with an expiry) - scanning
//      keys with no TTL is wasteful since they can never be expired.
//
//   2. Delete any sampled keys whose expiry time has passed.
//
//   3. Calculate what fraction of the sample was expired:
//        fraction = deleted / sampled
//
//   4. If fraction >= 0.25 (25% or more of the sample was expired), repeat
//      from step 1. A high hit rate means the keyspace is heavily polluted
//      with expired keys, so it is worth sampling again immediately.
//
//   5. If fraction < 0.25, stop and wait for the next 1s tick. The
//      keyspace is probably mostly clean - further sampling has diminishing
//      returns.
//
// WHY 25%?
// ---------
// It is a threshold. If 1 in 4 randomly sampled keys
// is already expired, the true expired fraction across the full keyspace is
// likely much higher - so we keep going. If fewer than 1 in 4 are expired,
// the keyspace is clean enough that the cost of another sample outweighs the
// memory we would reclaim.

package core

import (
	"log"
	"time"
)

func hasExpired(obj *Object) bool {
	exp, ok := expires[obj]
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func getExpiry(obj *Object) (uint64, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

func expireSample() float32 {
	var deleted int = 0
	var limit int = 20

	for key, obj := range store {
		exp, isExpirySet := getExpiry(obj)
		if isExpirySet {
			limit--

			if exp <= uint64(time.Now().UnixMilli()) {
				Del([]string{key})
				deleted++
			}
		}

		if limit == 0 {
			break
		}
	}

	return float32(deleted) / 20.0
}

func DeleteExpiredKeys() {
	for {
		fraction := expireSample()
		if fraction < 0.25 {
			break
		}
	}

	log.Println("deleted the expired but undeleted keys. total keys: ", len(store))
}
