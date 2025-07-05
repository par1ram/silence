module github.com/par1ram/silence/rpc/analytics

go 1.21

require (
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	github.com/influxdata/influxdb-client-go/v2 v2.12.0
	github.com/par1ram/silence/shared v0.0.0
)

replace github.com/par1ram/silence/shared => ../../shared
