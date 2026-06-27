# Kova

A Redis-compatible in-memory key-value store built from scratch in Go — no Redis libraries, no abstractions. Every layer from the TCP socket up through the eviction algorithm is hand-written.

Built as a deep-dive into how Redis actually works: event-driven I/O, RESP wire protocol, memory management, and persistence.

## Features

### Event-Driven I/O with Linux epoll
Rather than spawning a goroutine per connection (which doesn't scale past a few thousand clients), Kova uses a single event loop backed by Linux's `epoll` syscall. The server socket and every client socket are registered with epoll; the loop sleeps at `EpollWait` and wakes only when a file descriptor has data ready — zero CPU consumed while idle.

### RESP Protocol
Full hand-written encoder and decoder for the [Redis Serialization Protocol](https://redis.io/docs/reference/protocol-spec/), supporting all five types:

| Prefix | Type          | Example                          |
|--------|---------------|----------------------------------|
| `+`    | Simple string | `+OK\r\n`                        |
| `-`    | Error         | `-ERR unknown command\r\n`       |
| `:`    | Integer       | `:42\r\n`                        |
| `$`    | Bulk string   | `$5\r\nhello\r\n`                |
| `*`    | Array         | `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n` |

### Command Set

| Command | Syntax | Description |
|---------|--------|-------------|
| `PING` | `PING [message]` | Heartbeat |
| `SET` | `SET key value [EX seconds]` | Store a value with optional TTL |
| `GET` | `GET key` | Retrieve a value |
| `DEL` | `DEL key [key ...]` | Delete one or more keys |
| `TTL` | `TTL key` | Remaining time-to-live in seconds |
| `EXPIRE` | `EXPIRE key seconds` | Set a TTL on an existing key |
| `INCR` | `INCR key` | Atomically increment an integer |
| `INFO` | `INFO` | Keyspace statistics |
| `BGREWRITEAOF` | `BGREWRITEAOF` | Persist all keys to the AOF file |

### Command Pipelining
Multiple commands can be batched in a single client request. Kova decodes the full pipeline, evaluates every command, and flushes all responses in one `write` syscall.

### Object Model with Packed Type/Encoding
Each value is stored as an `Object` with its Redis type and encoding packed into a single `uint8` (4 bits each), matching how Redis represents objects internally. String values are auto-encoded:

- `INT` — value is a whole number
- `EMBSTR` — string ≤ 44 bytes  
- `RAW` — string > 44 bytes

The type system is defined for all five Redis types (String, List, Hash, Set, ZSet) with their corresponding encodings (LISTPACK, QUICKLIST, HASHTABLE, INTSET, SKIPLIST).

### Key Expiration
Two-layer expiry mirroring Redis:

- **Lazy** — expiry is checked on every `GET`; stale keys are deleted on access
- **Active** — a background cycle runs every second, samples up to 20 volatile keys, deletes expired ones, and loops again if ≥ 25% of the sample was stale (the same probabilistic threshold Redis uses)

### Memory Eviction
Three policies, configurable via `config/constants.go`:

| Policy | Behaviour |
|--------|-----------|
| `allkeys-lru` | Approximate LRU — samples 5 keys into a sorted eviction pool, evicts the least-recently-used first |
| `allkeys-random` | Evicts a random 40% of all keys |
| `simple-first` | Evicts whatever the first map iteration yields |

The LRU eviction pool stores `lastAccessedAt` timestamps as 24-bit wall-clock values (matching Redis's clock resolution) and sorts by idle time.

### AOF Persistence
`BGREWRITEAOF` serialises the entire keyspace to `kova-dump.aof` in RESP format. The AOF is also written automatically on graceful shutdown.

### Graceful Shutdown
Signal handler (`SIGINT` / `SIGTERM`) uses an atomic state machine with three states — `WAITING → BUSY → SHUTTING_DOWN` — to drain the current event-loop iteration before flushing the AOF and exiting. The `BUSY` state prevents a shutdown from racing with in-flight command processing.

---

## Getting Started

**Prerequisites:** Go 1.21+, Linux (epoll is Linux-only)

```bash
git clone https://github.com/Ozone317/Kova.git
cd Kova
go run main.go
```

The server listens on `127.0.0.1:7379`. Connect with `redis-cli`:

```bash
redis-cli -p 7379
```

```
127.0.0.1:7379> PING
PONG
127.0.0.1:7379> SET name "kova" EX 60
OK
127.0.0.1:7379> GET name
"kova"
127.0.0.1:7379> TTL name
(integer) 59
127.0.0.1:7379> INCR counter
(integer) 1
127.0.0.1:7379> INFO
# Keyspace
db0:keys=2,expires=1,avg_ttl=0
```

---

## Configuration

Edit `config/constants.go`:

```go
const (
    PORT                    = 7379
    HOST                    = "127.0.0.1"
    MAX_KEYS                = 1000          // trigger eviction above this
    AOF_FILE                = "./kova-dump.aof"
    EVICTION_POLICY         = "allkeys-lru" // allkeys-lru | allkeys-random | simple-first
    ALL_KEYS_EVICTION_RATIO = 0.4           // fraction of keys to evict per cycle
)
```

---

## Running Tests

```bash
make test
```

---

## License

MIT
