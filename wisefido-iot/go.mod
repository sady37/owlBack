module wisefido-iot

go 1.21

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.26.0
	owl-common v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace owl-common => ../owl-common
