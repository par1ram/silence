module github.com/par1ram/silence/api/auth

go 1.23.0

toolchain go1.23.2

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/par1ram/silence/shared v0.0.0
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.39.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.37.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/par1ram/silence/shared => ../../shared
