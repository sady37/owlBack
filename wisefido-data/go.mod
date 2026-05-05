module wisefido-data

go 1.25.0

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/xuri/excelize/v2 v2.10.0
	go.uber.org/zap v1.26.0
	golang.org/x/net v0.52.0
	gopkg.in/yaml.v3 v3.0.1
	owl-common v0.0.0
	wisefido-qinglan v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.4 // indirect
	github.com/tiendc/go-deepcopy v1.7.1 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)

replace owl-common => ../owl-common

replace wisefido-sensor => ../wisefido-sensor

replace wisefido-qinglan => ../wisefido-qinglan
