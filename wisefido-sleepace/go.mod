module wisefido-sleepace

go 1.25.0

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.16.1
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
	owl-common v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
)

replace owl-common => ../owl-common
