module github.com/par1ram/silence/rpc/dpi-bypass

go 1.21

require (
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/par1ram/silence/shared v0.0.0
)

replace github.com/par1ram/silence/shared => ../../shared
