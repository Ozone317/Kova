package config

const (
	PORT                    = 7379
	HOST                    = "127.0.0.1"
	MAX_KEYS                = 1000
	AOF_FILE                = "./kova-dump.aof"
	EVICTION_POLICY         = "allkeys-lru"
	ALL_KEYS_EVICTION_RATIO = 0.4
)
