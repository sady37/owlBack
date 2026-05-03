module wisefido-ai-health

go 1.25.0

require (
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
	owl-common v0.0.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace owl-common => ../owl-common
