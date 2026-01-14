module wisefido-radar

go 1.21

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.26.0
	owl-common v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eclipse/paho.mqtt.golang v1.4.3 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace owl-common => ../owl-common
