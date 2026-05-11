module wisefido-cardagg

go 1.25.0

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
	owl-common v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

replace owl-common => ../owl-common
