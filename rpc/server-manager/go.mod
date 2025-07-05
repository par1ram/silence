module github.com/par1ram/silence/rpc/server-manager

go 1.21

require (
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	github.com/docker/docker v24.0.5+incompatible
	github.com/par1ram/silence/shared v0.0.0
)

replace github.com/par1ram/silence/shared => ../../shared
